package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/publickey"
)

// newPublickeyItemSource は Publickey Adapter を組み立てる。
//
// @require httpClient != nil。
// @ensure 戻りは port.ItemSource。
func newPublickeyItemSource(httpClient *http.Client) port.ItemSource {
	return publickey.NewListItemSource(httpClient)
}
