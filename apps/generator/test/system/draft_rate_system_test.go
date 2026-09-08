//go:build system && ratemeasure

// Scope: System（原稿 API draft の prompt 精度 PASS 率計測。dispatch 専用。api で cursor / gemini を選ぶ）
// 実物: DRAFT_RATE_API=cursor なら cursorapi.TextWriter が実 TEST_CURSOR_API_KEY で Cloud Agents API（api.cursor.com）、
//
//	DRAFT_RATE_API=gemini なら geminiapi.TextWriter が実 TEST_GEMINI_API_KEY で generateContent
//	（generativelanguage.googleapis.com）を叩く。1 dispatch = 1 API（両方同時には測らない）。
//
// Double: なし。OAuth / Drive は呼ばない。fallback UseCase（manuscript.TextWriter）も経由しない。
// 目的: 固定擬似ソース → ComposeBriefWithTemplate(items, variant) → Write → ManuscriptDraftFromWriterOutput を
//
//	runs 回直列に通し、valid Draft が返る率を計測する（Decision 2026-09-03T14-47-00 / 2026-09-08T07-40-00）。
//	*cursorapi.Error / *geminiapi.Error の Op=="do"（API へ到達すらできない環境要因）はその回を分母から除外する。
//
// @require 選んだ API の key env（TEST_CURSOR_API_KEY / TEST_GEMINI_API_KEY）が process env にある（欠けたら Skip）。
//
//	本番 env 名（config.*APIKeyEnv）は読まない。Cursor CLI の `agent` binary は要らない（HTTP API 移行済み）。
//
// @ensure pass/(pass+fail) >= pass_threshold で緑、下回れば t.Fatalf。api・variant・文字数・環境 skip 回数を Logf。
// @invariant 既定 -tags=system では compile されない（ratemeasure tag）。local に secret を置かない。本番 key が計測へ流れない（TEST_ 直読み）。
package system

import (
	"context"
	_ "embed"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorapi"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/geminiapi"
)

//go:embed testdata/brief_prompt_variant_a.txt
var briefPromptVariantA string

// resolveBriefTemplate は DRAFT_PROMPT_VARIANT の値から brief template を返す。
// "default" は現行 constants.TextWriterBriefPrompt。それ以外は testdata の埋め込み variant。
func resolveBriefTemplate(t *testing.T, variant string) string {
	t.Helper()
	switch variant {
	case "", "default":
		return constants.TextWriterBriefPrompt
	case "a":
		return briefPromptVariantA
	default:
		t.Fatalf("DRAFT_PROMPT_VARIANT = %q, want default / a", variant)
		return ""
	}
}

// draftAPITarget は DRAFT_RATE_API が選ぶ 1 API 分の計測パラメタ。
type draftAPITarget struct {
	name      string // "cursor" | "gemini"（Logf 表示用）
	keyEnv    string // その API の test 用 key を読む env 名
	newWriter func(apiKey string) port.TextWriter
}

// resolveDraftAPITarget は DRAFT_RATE_API の値から計測対象 API を返す。
// "" / "cursor" は Cursor Cloud Agents、"gemini" は Gemini generateContent。
// why: 1 dispatch = 1 API（Decision 2026-09-08T07-40-00）。両方同時には測らない。
func resolveDraftAPITarget(t *testing.T, api string) draftAPITarget {
	t.Helper()
	switch api {
	case "", "cursor":
		return draftAPITarget{
			name:   "cursor",
			keyEnv: "TEST_CURSOR_API_KEY",
			newWriter: func(apiKey string) port.TextWriter {
				return cursorapi.NewTextWriter(&http.Client{}, apiKey)
			},
		}
	case "gemini":
		return draftAPITarget{
			name:   "gemini",
			keyEnv: "TEST_GEMINI_API_KEY",
			newWriter: func(apiKey string) port.TextWriter {
				return geminiapi.NewTextWriter(&http.Client{}, apiKey)
			},
		}
	default:
		t.Fatalf("DRAFT_RATE_API = %q, want cursor / gemini", api)
		return draftAPITarget{}
	}
}

// isEnvUnreachable は Write error が「API へ到達すらできなかった環境要因」かを返す。
// why: cursorapi.Error / geminiapi.Error は同型で、どちらも client.Do 失敗を Op:"do" で包む。
//
//	do 以降（応答が返った後）の失敗は prompt 精度の範疇なので分母に含める（Decision 2026-09-03T14-47-00）。
func isEnvUnreachable(err error) bool {
	var cerr *cursorapi.Error
	if errors.As(err, &cerr) {
		return cerr.Op == "do"
	}
	var gerr *geminiapi.Error
	if errors.As(err, &gerr) {
		return gerr.Op == "do"
	}
	return false
}

func draftEnvInt(t *testing.T, key string, def int) int {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		t.Fatalf("%s = %q, want 正の整数", key, raw)
	}
	return v
}

func draftEnvFloat(t *testing.T, key string, def float64) float64 {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 || v > 1 {
		t.Fatalf("%s = %q, want (0,1] の小数", key, raw)
	}
	return v
}

func TestDraftRate_measuresPassRate_overNRuns(t *testing.T) {
	// Given: DRAFT_RATE_API が選ぶ 1 API と、その API の実 key（欠けたら Skip = 環境要因、計測外）
	// 本番 env 名は読まない（Decision 2026-09-03T16-30-00）。
	api := strings.TrimSpace(os.Getenv("DRAFT_RATE_API"))
	target := resolveDraftAPITarget(t, api)
	apiKey := strings.TrimSpace(os.Getenv(target.keyEnv))
	if apiKey == "" {
		t.Skipf("計測 precondition: %s が無い（%s の draft rate 計測を skip）", target.keyEnv, target.name)
	}

	runs := draftEnvInt(t, "DRAFT_RATE_RUNS", 5)
	passThreshold := draftEnvFloat(t, "DRAFT_PASS_THRESHOLD", 0.8)
	variant := strings.TrimSpace(os.Getenv("DRAFT_PROMPT_VARIANT"))
	if variant == "" {
		variant = "default"
	}
	template := resolveBriefTemplate(t, variant)

	// Given: 固定擬似ソースから組んだ brief（seedSourceItems / draftTotalRunes は system_shared_test.go）
	brief, err := build.ComposeBriefWithTemplate(seedSourceItems(), template)
	if err != nil {
		t.Fatalf("ComposeBriefWithTemplate: %v", err)
	}

	// Given: 選んだ API の実 TextWriter（TEST_ から読んだ apiKey を直接渡す）。
	// why: Client.Timeout は置かない。1 呼び出しの全体上限は ctx。
	tw := target.newWriter(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	pass, fail, skipped := 0, 0, 0
	var sizes []int
	for i := 1; i <= runs; i++ {
		start := time.Now()
		raw, err := tw.Write(ctx, brief)
		elapsed := time.Since(start).Seconds()
		if err != nil {
			if isEnvUnreachable(err) {
				skipped++
				t.Logf("run %d/%d: 環境skip（Op=do: API へ到達不可: %v）所要 %.1fs", i, runs, err, elapsed)
				continue
			}
			fail++
			t.Logf("run %d/%d: FAIL（Write error: %v）所要 %.1fs", i, runs, err, elapsed)
			continue
		}
		draft, err := build.ManuscriptDraftFromWriterOutput(raw)
		if err != nil {
			fail++
			t.Logf("run %d/%d: FAIL（draft parse: %v）所要 %.1fs", i, runs, err, elapsed)
			continue
		}
		pass++
		total := draftTotalRunes(draft)
		sizes = append(sizes, total)
		t.Logf("run %d/%d: PASS（topic %d 件 / 全体 %d 文字 / 下限マージン %d）所要 %.1fs",
			i, runs, len(draft.Topics), total, total-constants.DraftTotalCharsMin, elapsed)
	}

	denom := pass + fail
	t.Logf("draft PASS率 %d/%d（環境skip %d）文字数 %v api=%s variant=%s", pass, denom, skipped, sizes, target.name, variant)

	if denom == 0 {
		t.Skipf("全 %d 回が環境要因（Op=do）で計測対象なし", runs)
	}
	rate := float64(pass) / float64(denom)
	if rate < passThreshold {
		t.Fatalf("draft PASS率 %.2f < 閾値 %.2f（%d/%d、環境skip %d、api=%s）", rate, passThreshold, pass, denom, skipped, target.name)
	}
}
