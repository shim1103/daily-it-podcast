package cloudwatch

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
// @ensure 結果は dc:date >= since を満たす item のみ。
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

	out := make([]models.SourceItem, 0, len(feed.Items))
	for _, item := range feed.Items {
		occurredAt, ok := parseDCDate(item.Date)
		if !ok || occurredAt.Before(since) {
			continue
		}
		out = append(out, toSourceItem(item, occurredAt))
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
	detail := models.SourceBody{Text: normalizeHTML(item.Description)}
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
