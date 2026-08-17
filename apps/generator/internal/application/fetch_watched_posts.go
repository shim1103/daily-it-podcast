package application

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

type FetchWatchedPosts struct {
	source port.PostSource
}

// @require source != nil
// @ensure 戻りは非 nil。
func NewFetchWatchedPosts(source port.PostSource) *FetchWatchedPosts {
	return &FetchWatchedPosts{source: source}
}

// @require uc != nil かつ uc.source != nil。now は since 算出の基準時刻。
// @ensure since は now.Add(-constants.FetchWindow)。WatchUserIDs 順に ListByUser し、各返却順を維持して 1 slice に結合する。
// @ensure ListByUser が最初に error を返した時点で即 return し、成功分の部分結果は返さない。以降の user は呼ばない。
// @invariant Infrastructure を参照しない。依存は port.PostSource と Entities のみ。
func (uc *FetchWatchedPosts) Run(ctx context.Context, now time.Time) ([]models.Post, error) {
	since := now.Add(-constants.FetchWindow)
	out := make([]models.Post, 0)
	for _, userID := range constants.WatchUserIDs {
		posts, err := uc.source.ListByUser(ctx, userID, since)
		if err != nil {
			return nil, err
		}
		out = append(out, posts...)
	}
	return out, nil
}
