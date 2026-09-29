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
	loc    *time.Location
}

// NewFetchSourceItems は表示 Location の昨日 half-open 窓で ItemSource を叩く UseCase を返す。
//
// @require source != nil かつ loc != nil。
// @ensure 戻りは非 nil。
func NewFetchSourceItems(source port.ItemSource, loc *time.Location) *FetchSourceItems {
	return &FetchSourceItems{source: source, loc: loc}
}

// Run は表示 Location の昨日 [since, until) を Adapter へ渡し、その結果を返す。
//
// @require uc != nil かつ uc.source != nil かつ uc.loc != nil。now は CLI 実行時刻（窓の暦日基準）。
// @ensure since, until は constants.YesterdayHalfOpenWindow(now, uc.loc)。source.List(ctx, since, until) を 1 回呼ぶ。
// @ensure List の戻りを再 filter せずそのまま返す（窓判定は Adapter の契約）。
// @ensure List が error ならその error を返し、成功結果は返さない。
// @invariant Infrastructure を参照しない。依存は port.ItemSource と Entities のみ。
func (uc *FetchSourceItems) Run(ctx context.Context, now time.Time) ([]models.SourceItem, error) {
	since, until := constants.YesterdayHalfOpenWindow(now, uc.loc)
	return uc.source.List(ctx, since, until)
}
