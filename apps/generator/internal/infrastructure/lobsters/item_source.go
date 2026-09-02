package lobsters

import (
	"context"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// apiBaseURL は Lobsters の base。
const apiBaseURL = "https://lobste.rs"

// 取得上限（この file が契約として値を固定する）。
//
// why: podcast は 1 日 1 回だけ生成する。上限値で 1 run あたりの fetch 回数を有界にする。
const (
	// MaxStoriesScanned は hottest.json 先頭から評価する story 数。
	MaxStoriesScanned = 25
	// MaxCommentsPerStory は 1 story あたり取得する comment 数の上限。
	MaxCommentsPerStory = 8
	// CommentDepth は取得する comment 階層の深さ（top-level のみ）。
	CommentDepth = 1
)

// ListItemSource は Lobsters の hottest を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は Lobsters 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// List は since 以降に発生した Lobsters story を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Context を key として解釈しない。
//
// 振る舞い（B/C が実装する）:
//   - hottest.json を取得し、created_at >= since の item だけ残し、先頭 MaxStoriesScanned 件の id を対象にする。
//   - 各 story について /s/<short_id>.json を取得し、comments[] を読む。
//   - deleted / moderated でない comment を先頭 MaxCommentsPerStory 件まで残し、comment_plain を本文に使う（CommentDepth=1）。
//   - Context は story の title/url と comment 本文から行を組む。
//   - OccurredAt は created_at（offset 付き RFC3339 相当）を UTC 化した値。
//   - SourceID = SourceID。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	panic("not implemented")
}
