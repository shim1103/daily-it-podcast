package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/cloudwatch"
)

// newCloudWatchItemSource は Impress クラウド Watch Adapter を組み立てる。
//
// @require httpClient != nil。
// @ensure 戻りは port.ItemSource。
func newCloudWatchItemSource(httpClient *http.Client) port.ItemSource {
	return cloudwatch.NewListItemSource(httpClient)
}
