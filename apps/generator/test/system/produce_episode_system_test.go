//go:build system

// Scope: System（e2e 1 回通し）
// 実物: composition.NewProduceEpisodeFromEnvWithTopicCount で結線した UseCase（本番と同じ結線だが、
//
//	topic 数だけ環境変数 SYSTEM_TEST_TOPIC_COUNT 由来）が、
//	実 5 情報源（HackerNews / Lobsters / Publickey / TechCrunch / クラウド Watch）→ 実 Cursor Cloud Agents API 原稿 →
//	実 Gemini TTS → 実 R2 書込 を 1 度だけ通す。
//
// Double: なし（test 専用 credential。実行場所は GHA）。R2_BUCKET は test 専用 bucket。
// 目的: 「system 全体が壊れていないか」を 1 回で見る（PASS 率は測らない。rate 計測は
//
//	tts_rate / draft_rate へ分離。Decision 2026-09-03T14-45-00）。
//	下位 Scope（HTTP / 配線 / schema 全 field）は再 assert しない。ここは orchestration の疎通だけ。
//
// @require process env に config 契約の全 key がある（1 つでも欠けたら Skip）。R2_BUCKET は test 専用 bucket。
//
//	Cursor CLI の `agent` binary は要らない（Cloud Agents HTTP API 移行済み。Decision 2026-09-03T17-03-33）。
//
// @ensure ProduceEpisode.Run が 1 回で完走する。成功時は episodeId を t.Log に出す。Fetch 窓内に SourceItem が 0 件だった日は
//
//	Domain Error（Op = no_source_items）で成功扱い（fetch は通っており system は壊れていない）。
//	それ以外の error は system 側の故障として t.Fatalf。
//
// @invariant local に secret を置かない。本番 credential / 本番 bucket を使わない。Run が書いた成果物の
//
//	cleanup はしない（同 stem upsert で残骸許容。Decision 2026-08-30T23-32-00）。
package system

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/composition"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

// systemTestTopicCountEnv は system-test 実行時に topic 数を注入する環境変数名。
const systemTestTopicCountEnv = "SYSTEM_TEST_TOPIC_COUNT"

// systemTestTopicCount は環境変数から topic 数を読む。未設定または parse 失敗時は
// 本番と同じ constants.DraftTopicCountTarget を使う（system-test を「topic 数を絞った
// 節約実行」に限定しない後方互換のデフォルト）。
func systemTestTopicCount() int {
	raw := strings.TrimSpace(os.Getenv(systemTestTopicCountEnv))
	if raw == "" {
		return constants.DraftTopicCountTarget
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return constants.DraftTopicCountTarget
	}
	return n
}

// systemConfigEnvKeys は ProduceEpisode を組むのに要る process env の全 key。
// 1 つでも空なら System 全体通しは実行できない。
var systemConfigEnvKeys = []string{
	config.CursorAPIKeyEnv,
	config.GeminiAPIKeyEnv,
	config.SpareGeminiAPIKeyEnv,
	config.R2AccessKeyIDEnv,
	config.R2SecretAccessKeyEnv,
	config.R2AccountIDEnv,
	config.R2BucketEnv,
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

	uc, err := composition.NewProduceEpisodeFromEnvWithTopicCount(systemTestTopicCount(), delivery.NewLogWriter(os.Stderr))
	if err != nil {
		t.Fatalf("NewProduceEpisodeFromEnvWithTopicCount: %v", err)
	}

	// ctx timeout: Cursor draft（数分）+ TTS topic+2 束（数分）+ R2 書込。余裕を持って 35 分。
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
