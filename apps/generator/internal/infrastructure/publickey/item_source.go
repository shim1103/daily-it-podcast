package publickey

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
// why: podcast は 1 日 1 回だけ生成する。window 内 entry の結果件数を有界にする。
const (
	// MaxStoriesScanned は結果 SourceItem に含める entry 数の上限。
	// published/updated >= since を満たした entry を、feed 先頭からこの件数まで集める。
	MaxStoriesScanned = 15
)

// ListItemSource は Publickey Atom feed を ItemSource として返す Adapter。
type ListItemSource struct {
	client   *http.Client
	maxItems int
	retry    port.RetryReporter
}

// NewListItemSource は Publickey 向け ItemSource を返す。
//
// @require httpClient != nil。retry != nil（Composition Root の結線責務）。
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
// @ensure maxItems <= 0 の場合、実効上限は MaxStoriesScanned にフォールバックする。
func NewListItemSource(httpClient *http.Client, maxItems int, retry port.RetryReporter) *ListItemSource {
	return &ListItemSource{client: httpClient, maxItems: maxItems, retry: retry}
}

// effectiveMaxStories は min(maxItems, MaxStoriesScanned) を実効上限として返す。
// maxItems <= 0 は呼び出し側の指定漏れとみなし、MaxStoriesScanned を安全側の値として使う。
func (s *ListItemSource) effectiveMaxStories() int {
	if s.maxItems <= 0 || s.maxItems > MaxStoriesScanned {
		return MaxStoriesScanned
	}
	return s.maxItems
}

// atomFeed は Atom feed のうち Adapter が使う field だけを表す。
type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

// atomEntry は entry のうち Adapter が使う field だけを表す。
type atomEntry struct {
	Title     string      `xml:"title"`
	ID        string      `xml:"id"`
	Published string      `xml:"published"`
	Updated   string      `xml:"updated"`
	Summary   string      `xml:"summary"`
	Content   atomContent `xml:"content"`
	Links     []atomLink  `xml:"link"`
	Author    atomAuthor  `xml:"author"`
}

type atomContent struct {
	Body string `xml:",chardata"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

// List は since 以降に発生した Publickey entry を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 結果は published（無ければ updated）>= since を満たす entry のみ。最大 min(maxItems, MaxStoriesScanned) 件（maxItems <= 0 は MaxStoriesScanned）。
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

	var feed atomFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, infraErr("decode_feed", err)
	}

	limit := s.effectiveMaxStories()
	out := make([]models.SourceItem, 0, limit)
	for _, entry := range feed.Entries {
		occurredAt, ok := parseEntryTime(entry)
		if !ok || occurredAt.Before(since) {
			continue
		}
		out = append(out, toSourceItem(entry, occurredAt))
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func parseEntryTime(entry atomEntry) (time.Time, bool) {
	for _, raw := range []string{entry.Published, entry.Updated} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			continue
		}
		return t.UTC(), true
	}
	return time.Time{}, false
}

func alternateHref(links []atomLink) string {
	for _, link := range links {
		if link.Rel == "alternate" && link.Href != "" {
			return link.Href
		}
	}
	return ""
}

func buildSummary(title, summary string) string {
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	switch {
	case title != "" && summary != "":
		return title + "\n" + summary
	case title != "":
		return title
	default:
		return summary
	}
}

// toSourceItem は Atom entry から SourceItem を組む。
// 写像の方針は Decision 2026-09-13T17-14-10。本 func が publickey 写像の正本。
func toSourceItem(entry atomEntry, occurredAt time.Time) models.SourceItem {
	detail := models.SourceBody{Text: httpget.NormalizeHTML(entry.Content.Body)}
	if href := alternateHref(entry.Links); href != "" {
		detail.Links = []string{href}
	}
	metaLines := make([]string, 0, 3)
	if entry.ID != "" {
		metaLines = append(metaLines, "item_id: "+entry.ID)
	}
	if name := strings.TrimSpace(entry.Author.Name); name != "" {
		metaLines = append(metaLines, "actor_id: "+name, "actor_name: "+name)
	}
	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: occurredAt,
		Summary:    buildSummary(entry.Title, entry.Summary),
		Detail:     detail,
		Discourse:  models.SourceBody{},
		Meta:       strings.Join(metaLines, "\n"),
	}
}

// getWithRetry は httpget へ委譲し、失敗を Adapter の infraErr で包む。
func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, err := httpget.GetWithRetry(ctx, s.client, url, s.retry, op)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}
