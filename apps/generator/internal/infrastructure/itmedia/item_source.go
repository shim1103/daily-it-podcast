package itmedia

import (
	"context"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// feedURL は取得対象の ITmedia NEWS feed。
//
// why: ITmedia NEWS 速報の RSS 2.0。1 outlet 1 adapter なので feed は 1 本に固定する。
const feedURL = "https://rss.itmedia.co.jp/rss/2.0/news_bursts.xml"

// ListItemSource は ITmedia NEWS 速報 feed を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は ITmedia NEWS 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// List は since 以降に発生した feed item を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Context を key として解釈しない。
//
// 振る舞い（B/C が実装する）:
//   - feedURL を GET し、encoding/xml で RSS 2.0 を parse する。
//   - pubDate >= since の item だけ残す。
//   - Context は item.title と item.description（token 効率のための軽い HTML 正規化のみ。full readability は行わない）
//     に "links: <item.link>" を加えた行で組む。
//   - OccurredAt は pubDate（RFC1123Z）を UTC 化した値。
//   - SourceID = SourceID。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	panic("not implemented")
}
