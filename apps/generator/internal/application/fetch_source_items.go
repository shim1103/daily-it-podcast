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
}

// @require source != nil
// @ensure 戻りは非 nil。
func NewFetchSourceItems(source port.ItemSource) *FetchSourceItems {
	return &FetchSourceItems{source: source}
}

// @require uc != nil かつ uc.source != nil。now は since 算出の基準時刻（CLI 実行時刻）。
// @ensure since は now.Add(-constants.FetchWindow)。source.List を 1 回呼ぶ。
// @ensure List が error ならその error を返し、成功結果は返さない。
// @invariant Infrastructure を参照しない。監視対象一覧を知らない。依存は port.ItemSource と Entities のみ。
func (uc *FetchSourceItems) Run(ctx context.Context, now time.Time) ([]models.SourceItem, error) {
	since := now.Add(-constants.FetchWindow)
	return uc.source.List(ctx, since)
}
