package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorapi"
)

// newCursorTextWriter は Cursor Cloud Agents API TextWriter Adapter を組み立てる。
//
// @require httpClient と cfg.APIKey は validation 済み。logw != nil。
// @ensure 戻りは port.TextWriter。
func newCursorTextWriter(httpClient *http.Client, cfg config.CursorConfig, logw *delivery.LogWriter) port.TextWriter {
	return cursorapi.NewTextWriter(httpClient, cfg.APIKey.Reveal(), logw)
}
