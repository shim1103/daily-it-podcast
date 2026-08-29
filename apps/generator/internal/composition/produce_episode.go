package composition

import (
	"os"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
)

// newProduceEpisode は検証済み Config から production 結線済みの全日次 UseCase を組み立てる。
//
// @require cfg は Generator の configuration boundary で検証済みである。
// @ensure 戻りは cfg の capability ごとに production Adapter を結線した *application.ProduceEpisode。
func newProduceEpisode(cfg config.Config) *application.ProduceEpisode {
	httpClient := sharedHTTPClient()
	fetch := application.NewFetchSourceItems(newGetXAPIItemSource(httpClient, cfg.Source))
	textWriter := newCursorTextWriter(cfg.Cursor)
	speech := newGeminiSpeechSynthesizer(httpClient, cfg.Gemini)
	writeEpisode := newGoogleDriveWriteEpisode(httpClient, cfg.Drive)
	return application.NewProduceEpisode(fetch, textWriter, speech, writeEpisode)
}

// NewProduceEpisodeFromEnv は process environment から Config を読み、production UseCase を組み立てる。
//
// @ensure config.Load が違反を返したら *config.Errors をそのまま返し、UseCase は nil。
// @invariant config.Load 呼び出しは Composition Root に閉じ、cmd / infrastructure へ漏らさない。
func NewProduceEpisodeFromEnv() (*application.ProduceEpisode, error) {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return nil, err
	}
	return newProduceEpisode(cfg), nil
}
