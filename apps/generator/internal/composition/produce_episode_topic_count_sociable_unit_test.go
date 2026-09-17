package composition

import (
	"io"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
)

// testConfigEnv は config.Load を通す最小の有効 process env（dummy 値）である。
// newProduceEpisodeWithTopicCount は結線だけを行い外部 I/O を起こさないため、
// 実 credential ではなく format 制約だけ満たす dummy 値で足りる。
var testConfigEnv = map[string]string{
	config.CursorAPIKeyEnv:      "dummy-cursor-api-key",
	config.GeminiAPIKeyEnv:      "dummy-gemini-api-key",
	config.SpareGeminiAPIKeyEnv: "dummy-spare-gemini-api-key",
	config.R2AccessKeyIDEnv:     "dummy-r2-access-key-id",
	config.R2SecretAccessKeyEnv: "dummy-r2-secret-access-key",
	config.R2AccountIDEnv:       "dummy-r2-account-id",
	config.R2BucketEnv:          "dummy-r2-bucket",
}

// newTestConfig は testConfigEnv を config.Load へ通し、検証済み Config を返す。
func newTestConfig(t *testing.T) config.Config {
	t.Helper()
	cfg, err := config.Load(func(key string) (string, bool) {
		v, ok := testConfigEnv[key]
		return v, ok
	})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return cfg
}

// newProduceEpisodeWithTopicCount は結線だけを行い外部 I/O を起こさないため、
// 検証済み Config を渡しても呼び出し自体で外部通信は発生しない。
// source constructor 5 関数（newHackerNewsItemSource 等）が topicCount を Adapter へ
// そのまま渡すことは、各 Adapter 実装の maxItems 実効性 test（例:
// internal/infrastructure/hackernews/item_source_sociable_unit_test.go の
// maxItems 打ち切り検証）が担う。ここでは composition 層の結線が topicCount を
// 受け取って構築できることだけを検証する（topicCount の実効性検証は
// internal/application 側の sociable unit test に委ねる）。

func TestNewProduceEpisodeWithTopicCount_returnsNonNilUseCase_whenTopicCountIsPositive(t *testing.T) {
	t.Parallel()

	// Given: 検証済み Config と任意の正 topicCount
	cfg := newTestConfig(t)
	logw := delivery.NewLogWriter(io.Discard)
	const topicCount = 3

	// When: newProduceEpisodeWithTopicCount を呼ぶ
	uc := newProduceEpisodeWithTopicCount(cfg, logw, topicCount)

	// Then: 非 nil の UseCase を返す（結線時点で外部 I/O は起きない）
	if uc == nil {
		t.Fatal("newProduceEpisodeWithTopicCount() = nil, want non-nil")
	}
}

func TestNewProduceEpisode_delegatesToWithTopicCountUsingDraftTopicCountTarget_whenCalled(t *testing.T) {
	t.Parallel()

	// Given: 検証済み Config
	cfg := newTestConfig(t)
	logw := delivery.NewLogWriter(io.Discard)

	// When: 本番用 newProduceEpisode を呼ぶ
	uc := newProduceEpisode(cfg, logw)

	// Then: newProduceEpisodeWithTopicCount(cfg, logw, constants.DraftTopicCountTarget) と同様に
	// 非 nil の UseCase を返す（本番 topicCount 固定の委譲経路が壊れていないことの smoke 検証）
	if uc == nil {
		t.Fatal("newProduceEpisode() = nil, want non-nil")
	}
}
