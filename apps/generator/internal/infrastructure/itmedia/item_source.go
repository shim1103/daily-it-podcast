package itmedia

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

// rssFeed は RSS 2.0 feed のうち Adapter が使う field だけを表す。
// package 外へ露出しない unexported 型。
type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// List は since 以降に発生した feed item を SourceItem slice で返す。
//
// @require since は OccurredAt の inclusive 下限。
// @ensure 各要素の SourceID は非空（= SourceID）。OccurredAt は UTC かつ since 以上。
// @ensure 該当なしは空 slice（nil ではない）。
// @invariant vendor 固有型・監視対象一覧を露出しない。Context を key として解釈しない。
func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("list", fmt.Errorf("client is nil"))
	}

	body, err := s.getWithRetry(ctx, feedURL, "fetch_feed")
	if err != nil {
		return nil, err
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, infraErr("decode_feed", err)
	}

	out := make([]models.SourceItem, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		occurredAt, ok := parsePubDate(item.PubDate)
		if !ok {
			continue
		}
		if occurredAt.Before(since) {
			continue
		}
		out = append(out, toSourceItem(item, occurredAt))
	}
	return out, nil
}

// parsePubDate は RFC1123Z 形式の pubDate を time.Time へ変換する。
func parsePubDate(pubDate string) (time.Time, bool) {
	t, err := time.Parse(time.RFC1123Z, strings.TrimSpace(pubDate))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// toSourceItem は RSS item から SourceItem を組む。
func toSourceItem(item rssItem, occurredAt time.Time) models.SourceItem {
	lines := make([]string, 0, 4)
	if item.Title != "" {
		lines = append(lines, "title: "+item.Title)
	}
	if text := normalizeHTML(item.Description); text != "" {
		lines = append(lines, "text: "+text)
	}
	if item.Link != "" {
		lines = append(lines, "permalink: "+item.Link, "links: "+item.Link)
	}
	return models.SourceItem{
		SourceID:   SourceID,
		OccurredAt: occurredAt.UTC(),
		Context:    strings.Join(lines, "\n"),
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

// get は GET を 1 回実行し body を返す。
// 2 つめの戻り値は再試行してよい失敗（client.Do error / 5xx）かどうか。
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

// pTagReplacer は <p> 段落区切りを改行へ変換する。
// why: normalizeHTML は 1 run で最大数百回呼ばれる。Replacer 生成を 1 回へ巻き上げる。
var pTagReplacer = strings.NewReplacer("<p>", "\n", "<P>", "\n")

// normalizeHTML は HTML entity を unescape し、<p> を改行へ、他タグを除去する軽い正規化。
// token 効率が目的で full readability はしない。
func normalizeHTML(s string) string {
	if s == "" {
		return ""
	}
	s = pTagReplacer.Replace(s)
	s = stripTags(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

// stripTags は "<" と ">" で囲まれた token を除去する。
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
