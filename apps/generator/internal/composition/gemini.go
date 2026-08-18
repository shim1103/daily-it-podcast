package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

// NewGeminiSpeechSynthesizer は Gemini TTS Adapter を組み立てる。
//
// @ensure 戻りは port.SpeechSynthesizer。
func NewGeminiSpeechSynthesizer() port.SpeechSynthesizer {
	return gemini.NewSpeechSynthesizer(&agentsecrets.Client{})
}
