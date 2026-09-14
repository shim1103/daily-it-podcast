package publickey

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*ListItemSource)(nil)

// ListItemSource は Publickey Atom feed を ItemSource として返す Adapter。
type ListItemSource struct {
	client *http.Client
}

// NewListItemSource は Publickey 向け ItemSource を返す。
//
// @require httpClient != nil
// @ensure 戻りは非 nil の *ListItemSource。vendor 固有型を露出しない。
func NewListItemSource(httpClient *http.Client) *ListItemSource {
	return &ListItemSource{client: httpClient}
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
// @ensure 結果は published（無ければ updated）>= since を満たす entry のみ。
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

	out := make([]models.SourceItem, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		occurredAt, ok := parseEntryTime(entry)
		if !ok || occurredAt.Before(since) {
			continue
		}
		out = append(out, toSourceItem(entry, occurredAt))
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
	detail := models.SourceBody{Text: normalizeHTML(entry.Content.Body)}
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

// getWithRetry は GET を実行し body を返す。
// client.Do error / 5xx は 1 回だけ即再試行（backoff なし）。4xx（429 含む）/ 読み取り失敗は即 return。
func (s *ListItemSource) getWithRetry(ctx context.Context, url, op string) ([]byte, error) {
	body, retryable, err := s.get(ctx, url)
	if err == nil {
		return body, nil
	}
	if !retryable {
		return nil, infraErr(op, err)
	}
	// why: 2 回目は retryable を問わず打ち切る（再試行は 1 回だけ）。
	body, _, err = s.get(ctx, url)
	if err != nil {
		return nil, infraErr(op, err)
	}
	return body, nil
}

func (s *ListItemSource) get(ctx context.Context, url string) (body []byte, retryable bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = res.Body.Close() }()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}
	if res.StatusCode >= 500 {
		return nil, true, fmt.Errorf("status %d", res.StatusCode)
	}
	if res.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("status %d", res.StatusCode)
	}
	return body, false, nil
}

var pTagReplacer = strings.NewReplacer("<p>", "\n", "<P>", "\n")

func normalizeHTML(s string) string {
	if s == "" {
		return ""
	}
	s = pTagReplacer.Replace(s)
	s = stripTags(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}
