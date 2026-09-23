package fetch

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// FetchSourceItems は取得窓を適用して ItemSource から SourceItem を取る UseCase である。
type FetchSourceItems struct {
	source port.ItemSource
}

// NewFetchSourceItems は ItemSource を包む Fetch UseCase を返す。
//
// @require source != nil
// @ensure 戻りは非 nil。
func NewFetchSourceItems(source port.ItemSource) *FetchSourceItems {
	return &FetchSourceItems{source: source}
}

// Run は now 基準の取得窓で source.List を 1 回呼ぶ。
//
// @require uc != nil かつ uc.source != nil。now は since 算出の基準時刻。
// @ensure since は now.Add(-constants.FetchWindow)。成功時は List の結果を返す。
// @ensure List が error ならその error を返し、成功結果は返さない。
// @invariant Infrastructure を参照しない。依存は port.ItemSource と Entities のみ。
func (uc *FetchSourceItems) Run(ctx context.Context, now time.Time) ([]models.SourceItem, error) {
	since := now.Add(-constants.FetchWindow)
	return uc.source.List(ctx, since)
}
