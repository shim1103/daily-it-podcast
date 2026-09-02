package itmedia_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/itmedia"
)

const feedPath = "/rss/2.0/news_bursts.xml"

// stubClientResponse は RoundTrip 1 回分の応答または error を表す。
type stubClientResponse struct {
	Status int
	Body   string
	Err    error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
// feedURL の path をキーに応答を返し、呼び出し回数を calls へ記録する。
type stubRoundTripper struct {
	byPath map[string]stubClientResponse
	calls  map[string]int
}

func newStubRoundTripper() *stubRoundTripper {
	return &stubRoundTripper{
		byPath: make(map[string]stubClientResponse),
		calls:  make(map[string]int),
	}
}

func (rt *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.URL.Path
	if key == "" {
		key = req.URL.String()
	}
	rt.calls[key]++

	res, ok := rt.byPath[key]
	if !ok {
		return nil, fmt.Errorf("stubRoundTripper: no response configured for %q", key)
	}
	if res.Err != nil {
		return nil, res.Err
	}
	status := res.Status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader([]byte(res.Body))),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func (rt *stubRoundTripper) setFeed(body string) {
	rt.byPath[feedPath] = stubClientResponse{Body: body}
}

func (rt *stubRoundTripper) setFeedResponse(res stubClientResponse) {
	rt.byPath[feedPath] = res
}

func (rt *stubRoundTripper) feedCalls() int {
	return rt.calls[feedPath]
}

func newStubListItemSource(rt *stubRoundTripper) *itmedia.ListItemSource {
	return itmedia.NewListItemSource(&http.Client{Transport: rt})
}

// rssItemFixture は RSS item XML を組むための入力。
type rssItemFixture struct {
	title       string
	link        string
	description string
	pubDate     string
}

func rssXML(items ...rssItemFixture) string {
	var b strings.Builder
	for _, it := range items {
		b.WriteString("<item>\n")
		if it.title != "" {
			fmt.Fprintf(&b, "<title>%s</title>\n", it.title)
		}
		if it.link != "" {
			fmt.Fprintf(&b, "<link>%s</link>\n", it.link)
		}
		if it.description != "" {
			if strings.Contains(it.description, "<") {
				fmt.Fprintf(&b, "<description><![CDATA[%s]]></description>\n", it.description)
			} else {
				fmt.Fprintf(&b, "<description>%s</description>\n", it.description)
			}
		}
		if it.pubDate != "" {
			fmt.Fprintf(&b, "<pubDate>%s</pubDate>\n", it.pubDate)
		}
		b.WriteString("</item>\n")
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<rss version="2.0">` + "\n" +
		"<channel>\n" + b.String() + "</channel>\n</rss>\n"
}

func formatRFC1123Z(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func TestList_parsesFeedAndMapsItemToSourceItem_whenItemInWindow(t *testing.T) {
	// @given feedURL の RSS 2.0 応答 double。window 内 item 1 件
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	occurredAt := since.Add(3 * time.Hour)
	pubDate := formatRFC1123Z(occurredAt)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(rssItemFixture{
		title:       "タイトル本文",
		link:        "https://www.itmedia.co.jp/news/articles/2609/02/article.html",
		description: "最初の段落<p>次の段落 <a href=\"https://e.example\">link</a> &amp; &lt;tag&gt;",
		pubDate:     pubDate,
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
	if got[0].SourceID != itmedia.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, itmedia.SourceID)
	}
	wantOccurredAt := occurredAt.UTC()
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
	wantContext := strings.Join([]string{
		"title: タイトル本文",
		"text: 最初の段落\n次の段落 link & <tag>",
		"permalink: https://www.itmedia.co.jp/news/articles/2609/02/article.html",
		"links: https://www.itmedia.co.jp/news/articles/2609/02/article.html",
	}, "\n")
	if got[0].Context != wantContext {
		t.Fatalf("Context =\n%q\nwant\n%q", got[0].Context, wantContext)
	}
}

func TestList_convertsJSTPubDateToUTC_whenOffsetIsPlus0900(t *testing.T) {
	// @given pubDate が JST (+0900) の feed double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	pubDate := "Wed, 02 Sep 2026 12:00:00 +0900"
	wantOccurredAt := time.Date(2026, 9, 2, 3, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(rssItemFixture{
		title:   "JST タイムゾーン",
		link:    "https://www.itmedia.co.jp/news/articles/jst.html",
		pubDate: pubDate,
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
	if !got[0].OccurredAt.Equal(wantOccurredAt) {
		t.Fatalf("OccurredAt = %v, want %v", got[0].OccurredAt, wantOccurredAt)
	}
	if got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt.Location() = %v, want UTC", got[0].OccurredAt.Location())
	}
}

func TestList_excludesItemsOlderThanSince_atBoundary(t *testing.T) {
	// @given pubDate == since の item と since より前の item を混ぜた double
	since := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(
		rssItemFixture{
			title:   "境界ちょうど",
			link:    "https://www.itmedia.co.jp/news/articles/boundary.html",
			pubDate: formatRFC1123Z(since),
		},
		rssItemFixture{
			title:   "境界の外",
			link:    "https://www.itmedia.co.jp/news/articles/old.html",
			pubDate: formatRFC1123Z(since.Add(-time.Second)),
		},
	))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
	if !strings.Contains(got[0].Context, "title: 境界ちょうど") {
		t.Fatalf("Context = %q, want boundary item only", got[0].Context)
	}
}

func TestList_buildsTitleOnlySourceItem_whenDescriptionEmpty(t *testing.T) {
	// @given description が空の item double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	occurredAt := since.Add(time.Hour)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(rssItemFixture{
		title:   "タイトルのみ",
		link:    "https://www.itmedia.co.jp/news/articles/title-only.html",
		pubDate: formatRFC1123Z(occurredAt),
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
	wantContext := strings.Join([]string{
		"title: タイトルのみ",
		"permalink: https://www.itmedia.co.jp/news/articles/title-only.html",
		"links: https://www.itmedia.co.jp/news/articles/title-only.html",
	}, "\n")
	if got[0].Context != wantContext {
		t.Fatalf("Context =\n%q\nwant\n%q", got[0].Context, wantContext)
	}
	if strings.Contains(got[0].Context, "text:") {
		t.Fatalf("Context = %q, want no text line", got[0].Context)
	}
}

func TestList_appendsLinksLineFromItemLink(t *testing.T) {
	// @given item.link を持つ feed double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	occurredAt := since.Add(2 * time.Hour)
	link := "https://www.itmedia.co.jp/news/articles/links-test.html"
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(rssItemFixture{
		title:       "リンク行テスト",
		link:        link,
		description: "本文あり",
		pubDate:     formatRFC1123Z(occurredAt),
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	wantLinks := "links: " + link
	if !strings.Contains(got[0].Context, wantLinks) {
		t.Fatalf("Context = %q, want %q", got[0].Context, wantLinks)
	}
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	// @given 全 item が since より古い double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(
		rssItemFixture{
			title:   "古い1",
			link:    "https://www.itmedia.co.jp/news/articles/old1.html",
			pubDate: formatRFC1123Z(since.Add(-time.Hour)),
		},
		rssItemFixture{
			title:   "古い2",
			link:    "https://www.itmedia.co.jp/news/articles/old2.html",
			pubDate: formatRFC1123Z(since.Add(-24 * time.Hour)),
		},
	))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("got = nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0 (%+v)", len(got), got)
	}
}

func TestList_returnsInfrastructureError_whenNon200OrInvalidXML(t *testing.T) {
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)

	t.Run("client nil", func(t *testing.T) {
		// @given client を持たない ListItemSource
		source := itmedia.NewListItemSource(nil)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *itmedia.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *itmedia.Error", err, err)
		}
		if !strings.HasPrefix(infra.Error(), "itmedia:") {
			t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "itmedia:")
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
		if infra.Op != "list" {
			t.Fatalf("Op = %q, want %q", infra.Op, "list")
		}
	})

	t.Run("5xx feed retries once", func(t *testing.T) {
		// @given feed が 500 を返し、再試行でも 500 を返す double
		rt := newStubRoundTripper()
		rt.setFeedResponse(stubClientResponse{Status: http.StatusInternalServerError, Body: "boom"})
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *itmedia.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *itmedia.Error", err, err)
		}
		if !strings.HasPrefix(infra.Error(), "itmedia:") {
			t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "itmedia:")
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
		if infra.Op != "fetch_feed" {
			t.Fatalf("Op = %q, want %q", infra.Op, "fetch_feed")
		}
		if rt.feedCalls() != 2 {
			t.Fatalf("feed fetch attempts = %d, want 2 (retry once)", rt.feedCalls())
		}
	})

	t.Run("4xx feed", func(t *testing.T) {
		// @given feed が 404 を返す double（再試行なし）
		rt := newStubRoundTripper()
		rt.setFeedResponse(stubClientResponse{Status: http.StatusNotFound, Body: "not found"})
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *itmedia.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *itmedia.Error", err, err)
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
		if infra.Op != "fetch_feed" {
			t.Fatalf("Op = %q, want %q", infra.Op, "fetch_feed")
		}
		if rt.feedCalls() != 1 {
			t.Fatalf("feed fetch attempts = %d, want 1 (no retry on 4xx)", rt.feedCalls())
		}
	})

	t.Run("invalid XML feed", func(t *testing.T) {
		// @given feed が壊れた XML を返す double
		rt := newStubRoundTripper()
		rt.setFeed("not-xml")
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *itmedia.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *itmedia.Error", err, err)
		}
		if !strings.HasPrefix(infra.Error(), "itmedia:") {
			t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "itmedia:")
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
		if infra.Op != "decode_feed" {
			t.Fatalf("Op = %q, want %q", infra.Op, "decode_feed")
		}
	})
}

func TestList_retriesOnceOnTransientError_whenSecondAttemptSucceeds(t *testing.T) {
	// @given feed が 1 回目 5xx、2 回目成功を返す double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	pubDate := "Wed, 02 Sep 2026 12:00:00 +0900"
	feedCalls := 0
	transientRT := &sequenceRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == feedPath {
				feedCalls++
				if feedCalls == 1 {
					return &http.Response{
						StatusCode: http.StatusBadGateway,
						Body:       io.NopCloser(strings.NewReader("boom")),
						Header:     make(http.Header),
						Request:    req,
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(rssXML(rssItemFixture{
						title:   "再試行成功",
						link:    "https://www.itmedia.co.jp/news/articles/retry.html",
						pubDate: pubDate,
					}))),
					Header:  make(http.Header),
					Request: req,
				}, nil
			}
			return nil, fmt.Errorf("unexpected path %q", req.URL.Path)
		},
	}
	source := itmedia.NewListItemSource(&http.Client{Transport: transientRT})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と feed fetch 回数
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if feedCalls != 2 {
		t.Fatalf("feed fetch attempts = %d, want 2 (retry once)", feedCalls)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
}

// sequenceRoundTripper は request ごとに任意の応答を返す最小 Transport。
type sequenceRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (rt *sequenceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.handler(req)
}
