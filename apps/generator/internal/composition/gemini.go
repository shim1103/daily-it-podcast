package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

// newGeminiSpeechSynthesizerPrimary は Gemini TTS Adapter を、
// 本番 free 枠（cfg.APIKey）の primary source として組み立てる。
//
// @require httpClient != nil。cfg.APIKey は検証済み。logw != nil。
// @ensure 戻りは port.SpeechSynthesizer。TierFree で構築する。TTS 1 呼び出しの長い timeout は Adapter が付け直す。
func newGeminiSpeechSynthesizerPrimary(httpClient *http.Client, cfg config.GeminiConfig, logw *delivery.LogWriter) port.SpeechSynthesizer {
	return gemini.NewSpeechSynthesizer(httpClient, cfg.APIKey.Reveal(), gemini.TierFree, logw)
}

// newGeminiSpeechSynthesizerSpare は Gemini TTS Adapter を、
// paid 枠（cfg.SpareAPIKey）の final-fallback source として組み立てる。
//
// @require httpClient != nil。cfg.SpareAPIKey は検証済み。logw != nil。
// @ensure 戻りは port.SpeechSynthesizer。TierPaid で構築する。TTS 1 呼び出しの長い timeout は Adapter が付け直す。
func newGeminiSpeechSynthesizerSpare(httpClient *http.Client, cfg config.GeminiConfig, logw *delivery.LogWriter) port.SpeechSynthesizer {
	return gemini.NewSpeechSynthesizer(httpClient, cfg.SpareAPIKey.Reveal(), gemini.TierPaid, logw)
}
