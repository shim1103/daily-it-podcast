package port

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// PostSource は監視 user の時間窓内オリジナル投稿を取得する。
// vendor HTTP・ページング・raw の振り分けは Infrastructure に閉じる。
//
// @require userID は空でない X user id。since は CreatedAt の inclusive 下限。
// @ensure userID のオリジナル投稿かつ CreatedAt >= since のみ返す。Reply / Repost / 引用 Repost は含めない。該当なしは空 slice。
// @invariant vendor 固有型を露出しない。添付が無い・取れない場合 Media は空でよい。
type PostSource interface {
	ListByUser(ctx context.Context, userID string, since time.Time) ([]models.Post, error)
}
