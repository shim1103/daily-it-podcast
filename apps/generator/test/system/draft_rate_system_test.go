//go:build system && ratemeasure

// Scope: System（CursorCLI draft の prompt 精度 PASS 率計測。dispatch 専用）
// 実物: cursorcli.TextWriter が実 CURSOR_API_KEY で実 Cursor CLI（agent）を叩く。
// Double: なし。GetX / Gemini / OAuth / Drive は呼ばない。
// 目的: 固定擬似ソース → ComposeBriefWithTemplate(items, variant) → Write → ManuscriptDraftFromWriterOutput を
//
//	runs 回直列に通し、valid Draft が返る率を計測する（Decision 2026-09-03T14-47-00）。
//	*cursorcli.Error の Op=="run"（環境経由）はその回を分母から除外する。
//
// @require process env に TEST_CURSOR_API_KEY がある / `agent` が PATH で解決できる（欠けたら Skip）。本番 env 名（config.CursorAPIKeyEnv）は読まない。
// @ensure pass/(pass+fail) >= pass_threshold で緑、下回れば t.Fatalf。variant・文字数・環境 skip 回数を Logf。
// @invariant 既定 -tags=system では compile されない（ratemeasure tag）。local に secret を置かない。本番 key が計測へ流れない（TEST_ 直読み）。
package system

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
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

func TestCursorCLIDraftRate_measuresPassRate_overNRuns(t *testing.T) {
	// Given: 実 TEST_CURSOR_API_KEY と PATH 上の `agent`（どちらか欠けたら Skip = 環境要因、計測外）
	// 本番 env 名は読まない（Decision 2026-09-03T16-30-00）。
	const cursorAPIKeyEnv = "TEST_CURSOR_API_KEY"
	apiKey := strings.TrimSpace(os.Getenv(cursorAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("計測 precondition: %s が無い（draft rate 計測を skip）", cursorAPIKeyEnv)
	}
	if _, err := exec.LookPath(cursorcli.BinaryName); err != nil {
		t.Skipf("計測 precondition: Cursor CLI %q が PATH で解決できない（draft rate 計測を skip）: %v", cursorcli.BinaryName, err)
	}

	runs := draftEnvInt(t, "DRAFT_RATE_RUNS", 5)
	passThreshold := draftEnvFloat(t, "DRAFT_PASS_THRESHOLD", 0.8)
	variant := strings.TrimSpace(os.Getenv("DRAFT_PROMPT_VARIANT"))
	if variant == "" {
		variant = "default"
	}
	template := resolveBriefTemplate(t, variant)

	// Given: 固定擬似ソースから組んだ brief（cursorcli_draft_system_test.go の seedSourceItems を流用）
	brief, err := build.ComposeBriefWithTemplate(seedSourceItems(), template)
	if err != nil {
		t.Fatalf("ComposeBriefWithTemplate: %v", err)
	}

	// why: factory の第1引数は「値」。TEST_ から読んだ apiKey をそのまま渡す。
	// factory が subprocess へ渡す env 名は cursorcli.CursorAPIKeyEnvName（CURSOR_API_KEY）のまま
	// （Cursor CLI 本体が CURSOR_API_KEY を期待するため。値の出所だけ TEST_ にする）。
	cursorFactory := processenv.NewSecretEnvLauncherFactory(apiKey, os.LookupEnv)
	tw := cursorcli.NewTextWriter(cursorFactory)

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	pass, fail, skipped := 0, 0, 0
	var sizes []int
	for i := 1; i <= runs; i++ {
		start := time.Now()
		raw, err := tw.Write(ctx, brief)
		elapsed := time.Since(start).Seconds()
		if err != nil {
			var cerr *cursorcli.Error
			if errors.As(err, &cerr) && cerr.Op == "run" {
				skipped++
				t.Logf("run %d/%d: 環境skip（Op=run: %v）所要 %.1fs", i, runs, err, elapsed)
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
	t.Logf("draft PASS率 %d/%d（環境skip %d）文字数 %v variant=%s", pass, denom, skipped, sizes, variant)

	if denom == 0 {
		t.Skipf("全 %d 回が環境要因（Op=run）で計測対象なし", runs)
	}
	rate := float64(pass) / float64(denom)
	if rate < passThreshold {
		t.Fatalf("draft PASS率 %.2f < 閾値 %.2f（%d/%d、環境skip %d）", rate, passThreshold, pass, denom, skipped)
	}
}
