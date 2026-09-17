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
// @ensure 戻りは port.SpeechSynthesizer。TTS 1 呼び出しの長い timeout は Adapter が付け直す。
func newGeminiSpeechSynthesizer(httpClient *http.Client, cfg config.GeminiConfig) port.SpeechSynthesizer {
	// why: free/paid 切り替えの結線は別 task。今回は TierFree を渡すだけ（Decision 2026-09-16T11-41-59）。
	return gemini.NewSpeechSynthesizer(httpClient, cfg.APIKey.Reveal(), gemini.TierFree)
}
