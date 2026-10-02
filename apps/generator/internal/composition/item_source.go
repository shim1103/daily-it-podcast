package composition

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"golang.org/x/sync/errgroup"
)

// compositeItemSource は複数の port.ItemSource を 1 Port に束ねる。
// 源個数を Application から隠す graph 組み立てであり、TextWriter / Speech の fallback 方針とは別責務。
// why: docs/decisions/2026-10-02T07-17-48
type compositeItemSource []port.ItemSource

// newCompositeItemSource は sources を束ねた合成 port.ItemSource を返す。
//
// @require 各 source は port.ItemSource 契約を満たす。sources は可変長で 0 本でもよい。
// @ensure 戻りは非 nil の port.ItemSource。
func newCompositeItemSource(sources ...port.ItemSource) port.ItemSource {
	return compositeItemSource(sources)
}

// List は各 source の List を並行に呼び、成功結果を連結して返す。
//
// @ensure 各 source を 1 回ずつ並行に呼ぶ。結果の連結順序は保証しない。
// @ensure いずれかの source.List が error を返したらその error を返し、成功分は返さない（他 source の実行は中断してよい）。
// @ensure 全 source が空、または source が 0 本のときも非 nil の空 slice を返す。
func (c compositeItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	slots, err := c.listSources(ctx, since)
	if err != nil {
		return nil, err
	}
	return concatSourceItemSlots(slots), nil
}

func (c compositeItemSource) listSources(ctx context.Context, since time.Time) ([][]models.SourceItem, error) {
	slots := make([][]models.SourceItem, len(c))
	g, gctx := errgroup.WithContext(ctx)
	// why: docs/decisions/2026-09-23T17-21-29
	g.SetLimit(5)

	for i, source := range c {
		g.Go(func() error {
			items, err := source.List(gctx, since)
			if err != nil {
				return err
			}
			slots[i] = items
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return slots, nil
}

func concatSourceItemSlots(slots [][]models.SourceItem) []models.SourceItem {
	merged := make([]models.SourceItem, 0)
	for _, items := range slots {
		merged = append(merged, items...)
	}
	return merged
}
