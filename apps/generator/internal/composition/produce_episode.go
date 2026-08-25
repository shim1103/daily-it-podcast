package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
)

// NewProduceEpisode は production 結線済みの全日次 UseCase を組み立てる。
//
// @ensure 戻りは production 結線済み *application.ProduceEpisode。
func NewProduceEpisode() *application.ProduceEpisode {
	fetch := application.NewFetchSourceItems(NewGetXAPIItemSource())
	textWriter := NewCursorTextWriter()
	speech := NewGeminiSpeechSynthesizer()
	writeEpisode := NewGoogleDriveWriteEpisode()
	return application.NewProduceEpisode(fetch, textWriter, speech, writeEpisode)
}
