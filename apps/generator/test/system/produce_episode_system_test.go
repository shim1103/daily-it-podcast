//go:build system

// Scope: System（e2e 1 回通し）
// 実物: composition.NewProduceEpisodeFromEnv で結線した本番 UseCase が、
//
//	実 3 情報源（HackerNews / Lobsters / ITmedia）→ 実 Cursor Cloud Agents API 原稿 →
//	実 Gemini TTS → 実 OAuth + Drive 書込 を 1 度だけ通す。
//
// Double: なし（test 専用 credential。実行場所は GHA）。DRIVE_FOLDER_ID は test 専用 folder。
// 目的: 「system 全体が壊れていないか」を 1 回で見る（PASS 率は測らない。rate 計測は
//
//	tts_rate / cursorapi_draft_rate へ分離。Decision 2026-09-03T14-45-00）。
//	下位 Scope（HTTP / 配線 / schema 全 field）は再 assert しない。ここは orchestration の疎通だけ。
//
// @require process env に config 契約の全 key がある（1 つでも欠けたら Skip）。DRIVE_FOLDER_ID は test 専用 folder。
//
//	Cursor CLI の `agent` binary は要らない（Cloud Agents HTTP API 移行済み。Decision 2026-09-03T17-03-33）。
//
// @ensure ProduceEpisode.Run が 1 回で完走する。成功時は episodeId を t.Log に出す。Fetch 窓内に SourceItem が 0 件だった日は
//
//	Domain Error（Op = no_source_items）で成功扱い（fetch は通っており system は壊れていない）。
//	それ以外の error は system 側の故障として t.Fatalf。
//
// @invariant local に secret を置かない。本番 credential / 本番 folder を使わない。Run が書いた成果物の
//
//	cleanup はしない（同 stem upsert で残骸許容。Decision 2026-08-30T23-32-00）。
package system

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/composition"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

// systemConfigEnvKeys は ProduceEpisode を組むのに要る process env の全 key。
// 1 つでも空なら System 全体通しは実行できない。
var systemConfigEnvKeys = []string{
	config.CursorAPIKeyEnv,
	config.GeminiAPIKeyEnv,
	config.GoogleOAuthClientIDEnv,
	config.GoogleOAuthClientSecretEnv,
	config.GoogleOAuthRefreshTokenEnv,
	config.DriveFolderIDEnv,
}

func requireSystemConfigEnv(t *testing.T) {
	t.Helper()
	var missing []string
	for _, key := range systemConfigEnvKeys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Skipf("System precondition: %s が無い（e2e 1 回通しを skip）", strings.Join(missing, " / "))
	}
}

func TestProduceEpisodeSystem_runsEndToEndOnce_whenAllCredentialsPresent(t *testing.T) {
	// Given: config 契約の全 key（1 つでも欠けたら Skip）
	requireSystemConfigEnv(t)

	uc, err := composition.NewProduceEpisodeFromEnv()
	if err != nil {
		t.Fatalf("NewProduceEpisodeFromEnv: %v", err)
	}

	// ctx timeout: Cursor draft（数分）+ TTS topic+2 束（数分）+ Drive 書込。余裕を持って 35 分。
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Minute)
	defer cancel()

	// When: production と同じ orchestration を 1 度だけ通す
	start := time.Now()
	episodeID, runErr := uc.Run(ctx, time.Now())
	elapsed := time.Since(start)

	// Then: 完走、または「Fetch 窓に SourceItem 0 件」の Domain Error のみ許す。
	if runErr == nil {
		t.Logf("e2e 1 回通し PASS（episodeId=%s 所要 %.1fs）", episodeID, elapsed.Seconds())
		return
	}
	var de *domainerrors.Error
	if errors.As(runErr, &de) && de.Op == domainerrors.OpNoSourceItems {
		t.Logf("e2e 1 回通し PASS（Fetch 窓内に SourceItem 0 件。fetch は疎通。所要 %.1fs）", elapsed.Seconds())
		return
	}
	t.Fatalf("ProduceEpisode.Run が失敗: %v（episodeId=%s 所要 %.1fs）", runErr, episodeID, elapsed.Seconds())
}
