package application

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

type FetchSourceItems struct {
	source port.ItemSource
	loc    *time.Location
}

// NewFetchSourceItems は表示 Location の昨日 half-open 窓で ItemSource を叩く UseCase を返す。
//
// @require source != nil かつ loc != nil。
// @ensure 戻りは非 nil。
func NewFetchSourceItems(source port.ItemSource, loc *time.Location) *FetchSourceItems {
	return &FetchSourceItems{source: source, loc: loc}
}

// Run は表示 Location の昨日 [since, until) に入る SourceItem を返す。
//
// @require uc != nil かつ uc.source != nil かつ uc.loc != nil。now は CLI 実行時刻（窓の暦日基準）。
// @ensure since, until は constants.YesterdayHalfOpenWindow(now, uc.loc)。source.List(ctx, since) を 1 回呼ぶ。
// @ensure 戻りは OccurredAt ∈ [since, until) の item のみ（Adapter が until 以降を返しても落とす）。
// @ensure List が error ならその error を返し、成功結果は返さない。
// @invariant Infrastructure を参照しない。監視対象一覧を知らない。依存は port.ItemSource と Entities のみ。
func (uc *FetchSourceItems) Run(ctx context.Context, now time.Time) ([]models.SourceItem, error) {
	since, until := constants.YesterdayHalfOpenWindow(now, uc.loc)
	items, err := uc.source.List(ctx, since)
	if err != nil {
		return nil, err
	}
	return filterOccurredInHalfOpen(items, since, until), nil
}

// filterOccurredInHalfOpen は OccurredAt ∈ [since, until) の要素だけを残す。
func filterOccurredInHalfOpen(items []models.SourceItem, since, until time.Time) []models.SourceItem {
	out := make([]models.SourceItem, 0, len(items))
	for _, it := range items {
		if !it.OccurredAt.Before(since) && it.OccurredAt.Before(until) {
			out = append(out, it)
		}
	}
	return out
}
