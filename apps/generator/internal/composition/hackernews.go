package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
)

// newHackerNewsItemSource は Hacker News Adapter を組み立てる。
//
// @require httpClient != nil。
// @ensure 戻りは port.ItemSource。
func newHackerNewsItemSource(httpClient *http.Client) port.ItemSource {
	return hackernews.NewListItemSource(httpClient)
}
