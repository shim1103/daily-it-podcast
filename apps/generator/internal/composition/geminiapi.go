package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/geminiapi"
)

// newGeminiTextWriterPrimary は Gemini generateContent TextWriter Adapter を、
// 本番 free 枠（cfg.APIKey）の primary source として組み立てる。
//
// @require httpClient と cfg.APIKey は validation 済み。logw != nil。
// @ensure 戻りは port.TextWriter。TierFree で構築する。logw を port.RetryReporter として渡す。
func newGeminiTextWriterPrimary(httpClient *http.Client, cfg config.GeminiConfig, logw *delivery.LogWriter) port.TextWriter {
	return geminiapi.NewTextWriter(httpClient, cfg.APIKey.Reveal(), geminiapi.TierFree, logw)
}

// newGeminiTextWriterSpare は Gemini generateContent TextWriter Adapter を、
// paid 枠（cfg.SpareAPIKey）の final-fallback source として組み立てる。
//
// @require httpClient と cfg.SpareAPIKey は validation 済み。logw != nil。
// @ensure 戻りは port.TextWriter。TierPaid で構築する。logw を port.RetryReporter として渡す。
func newGeminiTextWriterSpare(httpClient *http.Client, cfg config.GeminiConfig, logw *delivery.LogWriter) port.TextWriter {
	return geminiapi.NewTextWriter(httpClient, cfg.SpareAPIKey.Reveal(), geminiapi.TierPaid, logw)
}
