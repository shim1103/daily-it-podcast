package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
)

// newProduceEpisode は検証済み Config の capability ごとに production Adapter を結線した日次 UseCase を返す。
// 情報源は composite ItemSource 経由で束ね、Application へ情報源個数を渡さない。
//
// @require cfg は Generator の configuration boundary で検証済みである。
// @ensure 戻りは非 nil の *application.ProduceEpisode。
// @ensure Fetch は composite ItemSource 経由で行い、Application へ情報源個数を渡さない。
// @invariant config.Load 呼び出しをここで行わない。Composition Root の結線責務だけを持つ。
func newProduceEpisode(cfg config.Config) *application.ProduceEpisode {
	httpClient := sharedHTTPClient()
	fetch := application.NewFetchSourceItems(newCompositeItemSource(newGetXAPIItemSource(httpClient, cfg.Source)))
	textWriter := newCursorTextWriter(cfg.Cursor)
	speech := newGeminiSpeechSynthesizer(httpClient, cfg.Gemini)
	writeEpisode := newGoogleDriveWriteEpisode(httpClient, cfg.Drive)
	return application.NewProduceEpisode(fetch, textWriter, speech, writeEpisode, newEpisodeID, sharedDisplayLocation())
}

// NewProduceEpisodeFromEnv は process environment から Config を読み、production UseCase を組み立てる。
//
// @ensure config.Load が違反を返したら *config.Errors をそのまま返し、UseCase は nil。
// @invariant config.Load 呼び出しは Composition Root に閉じ、cmd / infrastructure へ漏らさない。
func NewProduceEpisodeFromEnv() (*application.ProduceEpisode, error) {
	cfg, err := config.Load(sharedLookupEnv())
	if err != nil {
		return nil, err
	}
	return newProduceEpisode(cfg), nil
}
