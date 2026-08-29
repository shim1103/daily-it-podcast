package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

// newGeminiSpeechSynthesizer は Gemini TTS Adapter を組み立てる。
//
// @require httpClient != nil。cfg は検証済み。
// @ensure 戻りは port.SpeechSynthesizer。
func newGeminiSpeechSynthesizer(httpClient *http.Client, cfg config.GeminiConfig) port.SpeechSynthesizer {
	return gemini.NewSpeechSynthesizer(httpClient, cfg.APIKey.Reveal())
}
