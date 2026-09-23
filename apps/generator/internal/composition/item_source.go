package composition

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// compositeItemSource は各 port.ItemSource.List を並行に呼び、結果を concat する。
//
// @invariant source 間で shared state を持たない（各 source は独立した port.ItemSource インスタンス）。
// @invariant source を跨いだ dedup / sort もしない（結果順序は保証しない）。
type compositeItemSource []port.ItemSource

// newCompositeItemSource は sources を登録順に束ねた合成 port.ItemSource を返す。
//
// @require 各 source は port.ItemSource 契約を満たす。sources は可変長で 0 本でもよい。
// @ensure 戻りは非 nil の port.ItemSource。
func newCompositeItemSource(sources ...port.ItemSource) port.ItemSource {
	return compositeItemSource(sources)
}

// List は各 source の List を並行に呼び、成功結果を連結して返す。
//
// @require since は OccurredAt の inclusive 下限（各 source の List 契約に委ねる）。
// @ensure 各 source を 1 回ずつ並行に呼ぶ。結果の連結順序は保証しない。
// @ensure いずれかの source.List が error を返したらその error を返し、成功分は返さない（他 source の実行は中断してよい）。
// @ensure 全 source が空、または source が 0 本のときも非 nil の空 slice を返す。
// @invariant vendor 固有型・情報源内部の監視対象一覧を露出しない。Summary / Detail / Discourse を key として解釈しない。
func (c compositeItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	merged := make([]models.SourceItem, 0)
	for _, source := range c {
		items, err := source.List(ctx, since)
		if err != nil {
			return nil, err
		}
		merged = append(merged, items...)
	}
	return merged, nil
}
