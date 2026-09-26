package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/publickey"
)

// newPublickeyItemSource は Publickey Adapter を組み立てる。
//
// @require httpClient != nil。
// @ensure 戻りは port.ItemSource。maxItems はそのまま Adapter へ渡す。
func newPublickeyItemSource(httpClient *http.Client, maxItems int, logw *delivery.LogWriter) port.ItemSource {
	return publickey.NewListItemSource(httpClient, maxItems, logw)
}
