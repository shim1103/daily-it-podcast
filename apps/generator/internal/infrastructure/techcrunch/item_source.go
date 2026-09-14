package techcrunch

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
	// pubDate >= since を満たした item を、feed 先頭からこの件数まで集める。
	MaxStoriesScanned = 15
)

// ListItemSource は TechCrunch RSS feed を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は TechCrunch 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
}

// rssFeed は RSS 2.0 feed のうち Adapter が使う field だけを表す。
type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

// rssItem は item のうち Adapter が使う field だけを表す。
// category / content:encoded は写像しない（捨てる）ため持たない。
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Creator     string `xml:"http://purl.org/dc/elements/1.1/ creator"`
}

// List は since 以降に発生した TechCrunch item を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 結果は pubDate >= since を満たす item のみ。最大 MaxStoriesScanned 件。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Summary / Detail / Discourse を key として解釈しない。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	body, err := s.getWithRetry(ctx, feedBaseURL, "fetch_feed")
	if err != nil {
		return nil, err
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, infraErr("decode_feed", err)
	}

	out := make([]models.SourceItem, 0, MaxStoriesScanned)
	for _, item := range feed.Channel.Items {
		occurredAt, ok := parsePubDate(item.PubDate)
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

func parsePubDate(pubDate string) (time.Time, bool) {
	t, err := time.Parse(time.RFC1123Z, strings.TrimSpace(pubDate))
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

// toSourceItem は RSS item から SourceItem を組む。
// 写像の方針は Decision 2026-09-13T17-14-10。本 func が techcrunch 写像の正本。
func toSourceItem(item rssItem, occurredAt time.Time) models.SourceItem {
	detail := models.SourceBody{Text: httpget.NormalizeHTML(item.Description)}
	if item.Link != "" {
		detail.Links = []string{item.Link}
	}
	metaLines := make([]string, 0, 3)
	if guid := strings.TrimSpace(item.GUID); guid != "" {
		metaLines = append(metaLines, "item_id: "+guid)
	}
	if creator := strings.TrimSpace(item.Creator); creator != "" {
		metaLines = append(metaLines, "actor_id: "+creator, "actor_name: "+creator)
	}
	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: occurredAt,
		Summary:    strings.TrimSpace(item.Title),
		Detail:     detail,
		Discourse:  models.SourceBody{},
		Meta:       strings.Join(metaLines, "\n"),
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
