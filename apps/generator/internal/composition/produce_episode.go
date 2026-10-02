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
// 情報源は HackerNews / Lobsters / Publickey / TechCrunch / クラウド Watch を選び、Application の CompositeItemSource へ渡して単一 ItemSource として Fetch する。
//
// @require cfg は Generator の configuration boundary で検証済みである。logw != nil。topicCount > 0。
// @ensure 戻りは非 nil の *application.ProduceEpisode。
// @ensure Fetch は Composition が選んだ Adapter 列を Application 合成（CompositeItemSource）経由で行い、FetchSourceItems へ情報源個数を渡さない。
// @invariant config.Load 呼び出しをここで行わない。Composition Root は結線と Adapter 構成だけを持つ（並行 List・fail-all・連結と fallback 方針は Application）。
func newProduceEpisodeWithTopicCount(cfg config.Config, logw *delivery.LogWriter, topicCount int) *application.ProduceEpisode {
	// why: logw は typed nil（*delivery.LogWriter の nil）になり得るが、port interface へ渡すと
	//      == nil 比較が false になり検知が漏れる（Go の typed nil 問題）。具体型の時点で唯一
	//      ここに fail-fast を集約する（Composition Root の結線責務）。
	if logw == nil {
		panic("composition: newProduceEpisodeWithTopicCount: logw is nil")
	}
	httpClient := appruntime.HTTPClient()
	displayLoc := appruntime.DisplayLocation()
	// Composition: どの情報源を何本選ぶか。振る舞い（並行 List）は Application CompositeItemSource。
	fetchUC := fetch.NewFetchSourceItems(
		newProductionItemSource(httpClient, sourceMaxItemsForTopicCount(topicCount), logw),
		displayLoc,
	)
	lookup := newR2CompletedEpisodeLookup(httpClient, cfg.R2, logw)
	// logw は port.FallbackReporter / port.ProgressReporter を満たす。application 用 callback の
	// 組み立ては delivery.LogWriter が持ち、Composition は結線だけ行う。
	// why: fallback 順序は GEMINI_API_KEY(free) → CURSOR_API_KEY → SPARE_GEMINI_API_KEY(paid, final)
	//      で固定する（Decision 2026-09-16T00-39-21）。
	textWriter := manuscript.NewTextWriter(
		[]port.TextWriter{
			newGeminiTextWriterPrimary(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini, logw),
			newCursorTextWriter(appruntime.HTTPClientWithoutTimeout(), cfg.Cursor, logw),
			newGeminiTextWriterSpare(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini, logw),
		},
		logw,
	)
	// why: TTS も TextWriter と同型の合成 layer 経由にする（Decision 2026-09-16T11-41-59）。
	//      fallback 順序は GEMINI_API_KEY(free) → SPARE_GEMINI_API_KEY(paid, final) で固定する
	//      （Decision 2026-09-16T00-39-21）。
	speech := speechapp.NewSpeechSynthesizer(
		[]port.SpeechSynthesizer{
			newGeminiSpeechSynthesizerPrimary(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini, logw),
			newGeminiSpeechSynthesizerSpare(appruntime.HTTPClientWithoutTimeout(), cfg.Gemini, logw),
		},
		logw,
	)
	encode := newFFmpegWAVToMP3Encoder()
	writeEpisode := newR2WriteEpisode(httpClient, cfg.R2, logw)
	return application.NewProduceEpisode(fetchUC, lookup, textWriter, speech, encode, writeEpisode, newEpisodeID, displayLoc, logw, topicCount)
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
