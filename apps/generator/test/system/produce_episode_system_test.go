//go:build system

// Scope: System（e2e 1 回通し）
// 実物: 本番結線の ProduceEpisode（topic 数のみ SYSTEM_TEST_TOPIC_COUNT）。実情報源 → 原稿 API → TTS → R2。
// Double: なし（test 専用 credential / R2_BUCKET。GHA）。
// @require config 契約の全 env がある（欠けたら Skip）。本番 credential / bucket を使わない。
// @ensure Run 前に test bucket の同日 date 一致 json（+ sibling mp3）を削除する。別日は触らない。
// @ensure 成功時 episodeID 非空。already_produced（ID 空の成功）は Fail。
// @ensure Op=no_source_items は成功扱い（Fetch 疎通）。それ以外の error は Fail。
// @ensure 成功 produce の直後 Run は already_produced（ID 空・error 無し）。
// @invariant 下位 Scope を再 assert しない。local に secret を置かない。
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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
	appruntime "github.com/shim1103/daily-it-podcast/apps/generator/internal/runtime"
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

	cfg, err := config.Load(appruntime.LookupEnv())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	now := time.Now()
	dateStr := now.In(appruntime.DisplayLocation()).Format("2006-01-02")
	logw := delivery.NewLogWriter(os.Stderr)
	cleaner := r2.NewSameDayObjectCleaner(
		appruntime.HTTPClient(),
		cfg.R2.AccessKeyID.Reveal(),
		cfg.R2.SecretAccessKey.Reveal(),
		cfg.R2.AccountID,
		cfg.R2.Bucket,
		logw,
	)
	removed, err := cleaner.RemoveForDate(context.Background(), dateStr)
	if err != nil {
		t.Fatalf("同日 object cleanup: %v", err)
	}
	t.Logf("同日 object cleanup: date=%s removed=%d", dateStr, removed)

	uc, err := composition.NewProduceEpisodeFromEnvWithTopicCount(systemTestTopicCount(), logw)
	if err != nil {
		t.Fatalf("NewProduceEpisodeFromEnvWithTopicCount: %v", err)
	}

	// what: draft + TTS + R2 の余裕（分単位）
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Minute)
	defer cancel()

	// When: 1 回目（full path）
	start := time.Now()
	episodeID, runErr := uc.Run(ctx, now)
	elapsed := time.Since(start)

	// Then
	if runErr == nil {
		if episodeID == "" {
			t.Fatalf("already_produced 相当（episodeID 空）。同日 cleanup 後に full path が通っていない（所要 %.1fs）", elapsed.Seconds())
		}
		t.Logf("e2e 1 回通し PASS（episodeId=%s 所要 %.1fs）", episodeID, elapsed.Seconds())

		// When: 2 回目（実 R2 上の同日冪等）
		id2, err2 := uc.Run(ctx, now)
		if err2 != nil {
			t.Fatalf("2 回目 Run が失敗: %v（episodeId=%s）", err2, id2)
		}
		if id2 != "" {
			t.Fatalf("2 回目 Run が再 produce した（episodeId=%s）。already_produced を期待", id2)
		}
		t.Logf("同日冪等 PASS（2 回目は already_produced）")
		return
	}
	var de *domainerrors.Error
	if errors.As(runErr, &de) && de.Op == domainerrors.OpNoSourceItems {
		t.Logf("e2e 1 回通し PASS（Fetch 窓内に SourceItem 0 件。fetch は疎通。所要 %.1fs）", elapsed.Seconds())
		return
	}
	t.Fatalf("ProduceEpisode.Run が失敗: %v（episodeId=%s 所要 %.1fs）", runErr, episodeID, elapsed.Seconds())
}
