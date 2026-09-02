package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/itmedia"
)

// newITmediaItemSource は ITmedia NEWS Adapter を組み立てる。
//
// @require httpClient != nil。
// @ensure 戻りは port.ItemSource。
func newITmediaItemSource(httpClient *http.Client) port.ItemSource {
	return itmedia.NewListItemSource(httpClient)
}
