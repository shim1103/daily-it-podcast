package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

// newGetXAPIItemSource は GetXAPI Adapter を組み立てる。
//
// @require httpClient != nil。cfg は検証済み。
// @ensure 戻りは port.ItemSource。
func newGetXAPIItemSource(httpClient *http.Client, cfg config.SourceConfig) port.ItemSource {
	return getxapi.NewPostSource(httpClient, cfg.GetXAPIKey.Reveal())
}
