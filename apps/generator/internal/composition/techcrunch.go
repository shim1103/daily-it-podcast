package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/techcrunch"
)

// newTechCrunchItemSource は TechCrunch Adapter を組み立てる。
//
// @require httpClient != nil。
// @ensure 戻りは port.ItemSource。maxItems はそのまま Adapter へ渡す。
func newTechCrunchItemSource(httpClient *http.Client, maxItems int) port.ItemSource {
	return techcrunch.NewListItemSource(httpClient, maxItems)
}
