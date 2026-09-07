//go:build system && ratemeasure

// Scope: System（Gemini generateContent 原稿 Adapter の実 API 疎通 smoke。dispatch 専用）
// 実物: geminiapi.TextWriter が実 TEST_GEMINI_API_KEY で実 generateContent（generativelanguage.googleapis.com）を 1 回叩く。
// Double: なし。Cursor / OAuth / Drive は呼ばない。fallback UseCase（manuscript.TextWriter）も経由しない。
// 目的: 最小の brief を Write へ 1 回渡し、Gemini から非空断片が返ることだけを確かめる（Decision 2026-09-07T19-06-00）。
//
//	PASS 率も尺も原稿品質も測らない。req が 1 往復して err == nil で断片が返れば緑。
//	retry / finishReason / 空 text の分類は Adapter の Sociable Unit / Narrow が固定済みなので
//	ここでは再確認しない。「実 endpoint・実 key で 200 と candidates が返る」配線だけを見る。
//	dispatch は generator-geminiapi-smoke.yml が SSOT。feature branch の geminiapi を叩くため
//	`gh workflow run generator-geminiapi-smoke.yml --ref <branch>` で回す（yml は master に置く）。
//
// @require process env に TEST_GEMINI_API_KEY がある（無ければ Skip = 環境要因、smoke 対象外）。
//
//	本番 env 名（config.GeminiAPIKeyEnv）は読まない（本番 key を smoke へ流さない）。
//
// @ensure Write が err == nil かつ trim 後非空の断片を返す。返らなければ t.Fatalf。所要秒と断片長を Logf。
// @invariant 既定 -tags=system では compile されない（ratemeasure tag）。local に secret を置かない。
//
//	master 単独では geminiapi package が無く compile fail する（feature ref でのみ緑。意図どおり）。
package system

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/geminiapi"
)

func TestGeminiAPISmoke_returnsFragment_overOneCall(t *testing.T) {
	// Given: 実 TEST_GEMINI_API_KEY（欠けたら Skip = 環境要因、smoke 対象外）
	const geminiAPIKeyEnv = "TEST_GEMINI_API_KEY"
	apiKey := strings.TrimSpace(os.Getenv(geminiAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("smoke precondition: %s が無い（geminiapi 疎通 smoke を skip）", geminiAPIKeyEnv)
	}

	// Given: 最小の brief（原稿品質は見ないので短い日本語 1 文で足りる）
	const brief = "「本日の IT ニュースです。」とだけ返してください。"

	// Given: 実 generateContent 経由の TextWriter（TEST_ から読んだ apiKey を直接渡す）。
	// why: Client.Timeout は置かない。1 呼び出しの全体上限は ctx。
	tw := geminiapi.NewTextWriter(&http.Client{}, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// When: Write を 1 回だけ通す
	start := time.Now()
	fragment, err := tw.Write(ctx, brief)
	elapsed := time.Since(start).Seconds()

	// Then: err == nil かつ非空断片（Gemini から req が 1 往復して返った）
	if err != nil {
		t.Fatalf("Write() error = %v（実 generateContent 疎通失敗）所要 %.1fs", err, elapsed)
	}
	if strings.TrimSpace(fragment) == "" {
		t.Fatalf("Write() が空断片を返した（疎通はしたが text が空）所要 %.1fs", elapsed)
	}
	t.Logf("geminiapi 疎通 OK（断片 %d 文字）所要 %.1fs", len([]rune(fragment)), elapsed)
}
