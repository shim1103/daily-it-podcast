package cloudwatch

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpget"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// 取得上限（この file が契約として値を固定する）。
//
// why: podcast は 1 日 1 回だけ生成する。window 内 item の結果件数を有界にする。
const (
	// MaxStoriesScanned は結果 SourceItem に含める item 数の上限。
	// dc:date >= since を満たした item を、feed 先頭からこの件数まで集める。
	MaxStoriesScanned = 15
)

// ListItemSource は Impress クラウド Watch RDF feed を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource はクラウド Watch 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// rdfFeed は RSS 1.0 / RDF feed のうち Adapter が使う field だけを表す。
// rdf:about / dc:subject / dc:creator は写像しない（捨てる）ため item に持たない。
type rdfFeed struct {
	Items []rdfItem `xml:"item"`
}

type rdfItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"http://purl.org/dc/elements/1.1/ date"`
}

// List は since 以降に発生したクラウド Watch item を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 結果は dc:date >= since を満たす item のみ。最大 MaxStoriesScanned 件。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Summary / Detail / Discourse を key として解釈しない。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	body, err := s.getWithRetry(ctx, feedURL, "fetch_feed")
	if err != nil {
		return nil, err
	}

	var feed rdfFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, infraErr("decode_feed", err)
	}

	out := make([]models.SourceItem, 0, MaxStoriesScanned)
	for _, item := range feed.Items {
		occurredAt, ok := parseDCDate(item.Date)
		if !ok || occurredAt.Before(since) {
			continue
		}
		out = append(out, toSourceItem(item, occurredAt))
		if len(out) >= MaxStoriesScanned {
			break
		}
	}
	return out, nil
}

func parseDCDate(raw string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

// toSourceItem は RDF item から SourceItem を組む。
// 写像の方針は Decision 2026-09-13T17-14-10。本 func が cloudwatch 写像の正本。
// Meta は空（作者を載せる契約なし）。rdf:about / dc:subject は捨てる。
func toSourceItem(item rdfItem, occurredAt time.Time) models.SourceItem {
	detail := models.SourceBody{Text: httpget.NormalizeHTML(item.Description)}
	if item.Link != "" {
		detail.Links = []string{item.Link}
	}
	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: occurredAt,
		Summary:    strings.TrimSpace(item.Title),
		Detail:     detail,
		Discourse:  models.SourceBody{},
		Meta:       "",
	}
}

// getWithRetry は httpget へ委譲し、失敗を Adapter の infraErr で包む。
func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, err := httpget.GetWithRetry(ctx, s.client, url)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}
