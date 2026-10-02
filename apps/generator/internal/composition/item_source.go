package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/fetch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// newProductionItemSource は本番の情報源 Adapter を選び、Application の合成型へ渡す。
// 並行 List・fail-all・連結の振る舞いは fetch.CompositeItemSource が持つ。
//
// @require httpClient != nil。
// @ensure 戻りは非 nil の port.ItemSource。
func newProductionItemSource(httpClient *http.Client, maxItems int) port.ItemSource {
	return fetch.NewCompositeItemSource([]port.ItemSource{
		newHackerNewsItemSource(httpClient, maxItems),
		newLobstersItemSource(httpClient, maxItems),
		newPublickeyItemSource(httpClient, maxItems),
		newTechCrunchItemSource(httpClient, maxItems),
		newCloudWatchItemSource(httpClient, maxItems),
	})
}

// sourceMaxItemsForTopicCount は Adapter へ渡す maxItems を決める。
// 本番既定 topicCount のときは 0（Adapter 既定へ委譲）。system-test で絞るときだけ件数を追従する。
func sourceMaxItemsForTopicCount(topicCount int) int {
	// why: maxItems <= 0 は各 Adapter が既存の MaxStoriesScanned へフォールバックする契約
	//      （infrastructure/*/item_source.go の effectiveMaxStories）。本番の topicCount は
	//      常に DraftTopicCountTarget なのでフォールバックへ委ね、既定値を変えない。
	//      system-test が topicCount を絞った時だけ、source 取得件数も追従して絞る。
	if topicCount != constants.DraftTopicCountTarget {
		return topicCount * constants.SourceItemsPerTopic
	}
	return 0
}
