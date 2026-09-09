package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/geminiapi"
)

// newGeminiTextWriter は Gemini generateContent TextWriter Adapter を組み立てる。
// Cursor の利用枠喪失時の原稿取得元（secondary）として manuscript UseCase へ渡す。
//
// @require httpClient と cfg.SpareAPIKey は validation 済み。
// @ensure 戻りは port.TextWriter。
func newGeminiTextWriter(httpClient *http.Client, cfg config.GeminiConfig) port.TextWriter {
	return geminiapi.NewTextWriter(httpClient, cfg.SpareAPIKey.Reveal())
}
