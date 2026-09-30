package port

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// ItemSource は時間窓内の情報を返す。
// vendor HTTP・監視対象一覧・Summary / Detail / Discourse / Meta の内部文法は Infrastructure に閉じる。
//
// @require since は OccurredAt の inclusive 下限。until は exclusive 上限。since.Before(until) であること。
// @ensure 各要素の SourceID は非空。OccurredAt は UTC かつ [since, until) に入る。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・情報源内部の監視対象一覧を露出しない。Summary / Detail / Discourse を key として解釈しない。
// @invariant 窓の両端判定は Adapter が行う。Application は再 filter しない。
type ItemSource interface {
	List(ctx context.Context, since, until time.Time) ([]models.SourceItem, error)
}
