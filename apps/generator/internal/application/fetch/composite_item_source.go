package fetch

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"golang.org/x/sync/errgroup"
)

// MaxConcurrentSourceLists は CompositeItemSource が同時に呼ぶ source.List の上限。
// why: docs/decisions/2026-09-23T17-21-29 — 情報源本数と揃えた安全側の初期値。
const MaxConcurrentSourceLists = 5

var _ port.ItemSource = (*CompositeItemSource)(nil)

// CompositeItemSource は複数の port.ItemSource を並行に List し結果を連結する。
// 源個数の知識は Composition が ctor へ渡す列に閉じ、FetchSourceItems は単一 Port として使う。
type CompositeItemSource struct {
	sources []port.ItemSource
}

// NewCompositeItemSource は sources を束ねた合成 port.ItemSource を返す。
//
// @require 各 source は port.ItemSource 契約を満たす。sources は nil または空でもよい。
// @ensure 戻りは非 nil の *CompositeItemSource。
func NewCompositeItemSource(sources []port.ItemSource) *CompositeItemSource {
	return &CompositeItemSource{sources: sources}
}

// List は各 source の List を並行に呼び、成功結果を連結して返す。
//
// @ensure 各 source を 1 回ずつ並行に呼ぶ。結果の連結順序は保証しない。
// @ensure いずれかの source.List が error を返したらその error を返し、成功分は返さない（他 source の実行は中断してよい）。
// @ensure 全 source が空、または source が 0 本のときも非 nil の空 slice を返す。
// @ensure 同時実行数は最大 MaxConcurrentSourceLists 件まで。
func (c *CompositeItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	slots, err := c.listSources(ctx, since)
	if err != nil {
		return nil, err
	}
	return concatSourceItemSlots(slots), nil
}

func (c *CompositeItemSource) listSources(ctx context.Context, since time.Time) ([][]models.SourceItem, error) {
	slots := make([][]models.SourceItem, len(c.sources))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(MaxConcurrentSourceLists)

	for i, source := range c.sources {
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
