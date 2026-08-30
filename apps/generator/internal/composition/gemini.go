package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

// newGeminiSpeechSynthesizer は Gemini TTS Adapter を組み立てる。
//
// @require cfg は検証済み。
// @ensure 戻りは port.SpeechSynthesizer。TTS 専用の長い HTTP timeout を使う。
func newGeminiSpeechSynthesizer(cfg config.GeminiConfig) port.SpeechSynthesizer {
	return gemini.NewSpeechSynthesizer(geminiHTTPClient(), cfg.APIKey.Reveal())
}
