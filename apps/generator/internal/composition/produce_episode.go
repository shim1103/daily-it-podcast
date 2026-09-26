package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/fetch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/manuscript"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	speechapp "github.com/shim1103/daily-it-podcast/apps/generator/internal/application/speech"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	appruntime "github.com/shim1103/daily-it-podcast/apps/generator/internal/runtime"
)

// newProduceEpisode は newProduceEpisodeWithTopicCount(cfg, logw, constants.DraftTopicCountTarget) へ委譲する。
// 本番経路は常に固定 topicCount を使う。
//
// @require cfg は Generator の configuration boundary で検証済みである。logw != nil。
// @ensure 戻りは非 nil の *application.ProduceEpisode。
func newProduceEpisode(cfg config.Config, logw *delivery.LogWriter) *application.ProduceEpisode {
	return newProduceEpisodeWithTopicCount(cfg, logw, constants.DraftTopicCountTarget)
}

// newProduceEpisodeWithTopicCount は検証済み Config の capability ごとに production Adapter を結線した日次 UseCase を返す。
// 情報源は HackerNews / Lobsters / Publickey / TechCrunch / クラウド Watch を composite ItemSource 経由で束ね、Application へ情報源個数を渡さない。
//
// @require cfg は Generator の configuration boundary で検証済みである。logw != nil。topicCount > 0。
// @ensure 戻りは非 nil の *application.ProduceEpisode。
// @ensure Fetch は composite ItemSource 経由で行い、Application へ情報源個数を渡さない。
// @invariant config.Load 呼び出しをここで行わない。Composition Root の結線責務だけを持つ。
func newProduceEpisodeWithTopicCount(cfg config.Config, logw *delivery.LogWriter, topicCount int) *application.ProduceEpisode {
	httpClient := appruntime.HTTPClient()
	// why: maxItems <= 0 は各 Adapter が既存の MaxStoriesScanned へフォールバックする契約
	//      （infrastructure/*/item_source.go の effectiveMaxStories）。本番の topicCount は
	//      常に DraftTopicCountTarget なのでフォールバックへ委ね、既定値を変えない。
	//      system-test が topicCount を絞った時だけ、source 取得件数も追従して絞る。
	sourceMaxItems := 0
	if topicCount != constants.DraftTopicCountTarget {
		sourceMaxItems = topicCount * constants.SourceItemsPerTopic
	}
	fetchUC := fetch.NewFetchSourceItems(newCompositeItemSource(
		newHackerNewsItemSource(httpClient, sourceMaxItems),
		newLobstersItemSource(httpClient, sourceMaxItems),
		newPublickeyItemSource(httpClient, sourceMaxItems),
		newTechCrunchItemSource(httpClient, sourceMaxItems),
		newCloudWatchItemSource(httpClient, sourceMaxItems),
	))
	lookup := newR2CompletedEpisodeLookup(httpClient, cfg.R2)
	// logw は port.FallbackReporter / port.ProgressReporter を満たす。application 用 callback の
	// 組み立ては delivery.LogWriter が持ち、Composition は結線だけ行う。
	// why: fallback 順序は GEMINI_API_KEY(free) → CURSOR_API_KEY → SPARE_GEMINI_API_KEY(paid, final)
	//      で固定する（Decision 2026-09-16T00-39-21）。
	textWriter := manuscript.NewTextWriter(
		[]port.TextWriter{
			newGeminiTextWriterPrimary(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini),
			newCursorTextWriter(appruntime.HTTPClientWithoutTimeout(), cfg.Cursor),
			newGeminiTextWriterSpare(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini),
		},
		logw,
	)
	// why: TTS も TextWriter と同型の合成 layer 経由にする（Decision 2026-09-16T11-41-59）。
	//      fallback 順序は GEMINI_API_KEY(free) → SPARE_GEMINI_API_KEY(paid, final) で固定する
	//      （Decision 2026-09-16T00-39-21）。
	speech := speechapp.NewSpeechSynthesizer(
		[]port.SpeechSynthesizer{
			newGeminiSpeechSynthesizerPrimary(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini),
			newGeminiSpeechSynthesizerSpare(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini),
		},
		logw,
	)
	encode := newFFmpegWAVToMP3Encoder()
	writeEpisode := newR2WriteEpisode(httpClient, cfg.R2)
	return application.NewProduceEpisode(fetchUC, lookup, textWriter, speech, encode, writeEpisode, newEpisodeID, appruntime.DisplayLocation(), logw, topicCount)
}

// NewProduceEpisodeFromEnv は process environment から Config を読み、
// constants.DraftTopicCountTarget（本番固定値）で production UseCase を組み立てる。
//
// @require logw != nil。
// @ensure config.Load が違反を返したら *config.Errors をそのまま返し、UseCase は nil。
// @invariant config.Load 呼び出しは Composition Root に閉じ、cmd / infrastructure へ漏らさない。
func NewProduceEpisodeFromEnv(logw *delivery.LogWriter) (*application.ProduceEpisode, error) {
	return NewProduceEpisodeFromEnvWithTopicCount(constants.DraftTopicCountTarget, logw)
}

// NewProduceEpisodeFromEnvWithTopicCount は process environment から Config を読み、
// 任意の topicCount で production UseCase を組み立てる。system-test が
// SYSTEM_TEST_TOPIC_COUNT 経由で topic 数を絞った実行を行うための入口であり、
// 本番経路（NewProduceEpisodeFromEnv）は常に constants.DraftTopicCountTarget を渡す。
//
// @require logw != nil。topicCount > 0。
// @ensure config.Load が違反を返したら *config.Errors をそのまま返し、UseCase は nil。
// @invariant config.Load 呼び出しは Composition Root に閉じ、cmd / infrastructure へ漏らさない。
func NewProduceEpisodeFromEnvWithTopicCount(topicCount int, logw *delivery.LogWriter) (*application.ProduceEpisode, error) {
	cfg, err := config.Load(appruntime.LookupEnv())
	if err != nil {
		return nil, err
	}
	return newProduceEpisodeWithTopicCount(cfg, logw, topicCount), nil
}
