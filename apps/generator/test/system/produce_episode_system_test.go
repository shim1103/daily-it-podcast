//go:build system

// Scope: System（e2e 1 回通し）
// 実物: 本番結線の ProduceEpisode（topic 数のみ SYSTEM_TEST_TOPIC_COUNT）。実情報源 → 原稿 API → TTS → R2。
// Double: なし（test 専用 credential / R2_BUCKET。GHA）。
// @require config 契約の全 env がある（欠けたら Skip）。本番 credential / bucket を使わない。
// @ensure Run が完走する、または Op=no_source_items で成功扱い。それ以外は Fail。
// @invariant 下位 Scope を再 assert しない。成果物 cleanup しない。local に secret を置かない。
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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

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
	// Given
	requireSystemConfigEnv(t)

	uc, err := composition.NewProduceEpisodeFromEnvWithTopicCount(systemTestTopicCount(), delivery.NewLogWriter(os.Stderr))
	if err != nil {
		t.Fatalf("NewProduceEpisodeFromEnvWithTopicCount: %v", err)
	}

	// what: draft + TTS + R2 の余裕（分単位）
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Minute)
	defer cancel()

	// When
	start := time.Now()
	episodeID, runErr := uc.Run(ctx, time.Now())
	elapsed := time.Since(start)

	// Then
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
