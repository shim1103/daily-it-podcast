package techcrunch_test

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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/techcrunch"
)

const feedPath = "/feed/"

// stubClientResponse は RoundTrip 1 回分の応答または error を表す。
type stubClientResponse struct {
	Status int
	Body   string
	Err    error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
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
	path := req.URL.Path
	rt.calls[path]++
	res, ok := rt.byPath[path]
	if !ok {
		return nil, fmt.Errorf("stubRoundTripper: no response configured for path %q", path)
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

func newStubListItemSource(rt *stubRoundTripper) *techcrunch.ListItemSource {
	return techcrunch.NewListItemSource(&http.Client{Transport: rt})
}

// rssItemFixture は RSS item XML を組むための入力。
type rssItemFixture struct {
	title          string
	link           string
	description    string
	pubDate        string
	guid           string
	creator        string
	categories     []string
	contentEncoded string
}

func rssXML(items ...rssItemFixture) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:dc="http://purl.org/dc/elements/1.1/">` + "\n")
	b.WriteString("<channel>\n<title>TechCrunch</title>\n")
	for _, it := range items {
		b.WriteString("<item>\n")
		if it.title != "" {
			fmt.Fprintf(&b, "<title>%s</title>\n", it.title)
		}
		if it.link != "" {
			fmt.Fprintf(&b, "<link>%s</link>\n", it.link)
		}
		if it.creator != "" {
			fmt.Fprintf(&b, "<dc:creator><![CDATA[%s]]></dc:creator>\n", it.creator)
		}
		if it.pubDate != "" {
			fmt.Fprintf(&b, "<pubDate>%s</pubDate>\n", it.pubDate)
		}
		for _, cat := range it.categories {
			fmt.Fprintf(&b, "<category><![CDATA[%s]]></category>\n", cat)
		}
		if it.guid != "" {
			fmt.Fprintf(&b, `<guid isPermaLink="false">%s</guid>`+"\n", it.guid)
		}
		if it.description != "" {
			fmt.Fprintf(&b, "<description><![CDATA[%s]]></description>\n", it.description)
		}
		if it.contentEncoded != "" {
			fmt.Fprintf(&b, "<content:encoded><![CDATA[%s]]></content:encoded>\n", it.contentEncoded)
		} else {
			b.WriteString("<content:encoded><![CDATA[]]></content:encoded>\n")
		}
		b.WriteString("</item>\n")
	}
	b.WriteString("</channel>\n</rss>\n")
	return b.String()
}

func formatRFC1123Z(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func TestList_mapsRSSItemToSourceItem_whenItemInWindow(t *testing.T) {
	// @given RSS 2.0 feed double。window 内 item 1 件（guid / dc:creator / 空 content:encoded / category あり）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	occurredAt := since.Add(3 * time.Hour)
	articleURL := "https://techcrunch.com/2026/09/01/example-article/"
	guid := "https://techcrunch.com/?p=3163457"
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(rssItemFixture{
		title:       "タイトル本文",
		link:        articleURL,
		description: "説明テキスト &amp; more",
		pubDate:     formatRFC1123Z(occurredAt),
		guid:        guid,
		creator:     "Connie Loizos",
		categories:  []string{"AI", "Startups"},
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と写像
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
	if got[0].SourceID != techcrunch.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, techcrunch.SourceID)
	}
	wantOccurredAt := occurredAt.UTC()
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
	if got[0].Summary != "タイトル本文" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "タイトル本文")
	}
	if got[0].Detail.Text != "説明テキスト & more" {
		t.Fatalf("Detail.Text = %q, want description text", got[0].Detail.Text)
	}
	if len(got[0].Detail.Links) != 1 || got[0].Detail.Links[0] != articleURL {
		t.Fatalf("Detail.Links = %#v, want article URL", got[0].Detail.Links)
	}
	if !got[0].Discourse.Empty() {
		t.Fatalf("Discourse = %+v, want empty", got[0].Discourse)
	}
	if !strings.Contains(got[0].Meta, "item_id: "+guid) {
		t.Fatalf("Meta = %q, want guid as item_id", got[0].Meta)
	}
	if strings.Contains(got[0].Meta, articleURL) {
		t.Fatalf("Meta = %q, must not use article URL as item_id", got[0].Meta)
	}
	if !strings.Contains(got[0].Meta, "actor_id: Connie Loizos") || !strings.Contains(got[0].Meta, "actor_name: Connie Loizos") {
		t.Fatalf("Meta = %q, want dc:creator as actor", got[0].Meta)
	}
	if strings.Contains(got[0].Summary, "AI") || strings.Contains(got[0].Detail.Text, "AI") || strings.Contains(got[0].Meta, "AI") {
		t.Fatalf("category leaked into SourceItem: Summary=%q Detail=%q Meta=%q", got[0].Summary, got[0].Detail.Text, got[0].Meta)
	}
}

func TestList_excludesItemsOlderThanSince_atBoundary(t *testing.T) {
	// @given pubDate==since の item と pubDate==since-1s の item を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(
		rssItemFixture{
			title:   "境界ちょうど",
			link:    "https://techcrunch.com/in/",
			pubDate: formatRFC1123Z(since),
			guid:    "https://techcrunch.com/?p=1",
		},
		rssItemFixture{
			title:   "境界の外",
			link:    "https://techcrunch.com/out/",
			pubDate: formatRFC1123Z(since.Add(-time.Second)),
			guid:    "https://techcrunch.com/?p=2",
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
	if got[0].Summary != "境界ちょうど" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "境界ちょうど")
	}
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	// @given 全 item が since より古い double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(
		rssItemFixture{
			title:   "古い1",
			link:    "https://techcrunch.com/old1/",
			pubDate: formatRFC1123Z(since.Add(-time.Hour)),
			guid:    "https://techcrunch.com/?p=10",
		},
		rssItemFixture{
			title:   "古い2",
			link:    "https://techcrunch.com/old2/",
			pubDate: formatRFC1123Z(since.Add(-24 * time.Hour)),
			guid:    "https://techcrunch.com/?p=11",
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

func TestList_returnsInfrastructureError_whenClientNilOrNon200OrInvalidXML(t *testing.T) {
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("client nil", func(t *testing.T) {
		// @given client を持たない ListItemSource
		source := techcrunch.NewListItemSource(nil)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		assertTechCrunchInfraError(t, err)
	})

	t.Run("non-200 feed", func(t *testing.T) {
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
		assertTechCrunchInfraError(t, err)
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
		assertTechCrunchInfraError(t, err)
	})
}

func TestList_retriesOnceOnTransientError_whenSecondAttemptSucceeds(t *testing.T) {
	// @given feed が 1 回目 5xx、2 回目成功を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	pubDate := formatRFC1123Z(since.Add(time.Hour))
	feedCalls := 0
	transientRT := &sequenceRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			feedCalls++
			if feedCalls == 1 {
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Body:       io.NopCloser(strings.NewReader("boom")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			body := rssXML(rssItemFixture{
				title:   "再試行成功",
				link:    "https://techcrunch.com/retry/",
				pubDate: pubDate,
				guid:    "https://techcrunch.com/?p=99",
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}
	source := techcrunch.NewListItemSource(&http.Client{Transport: transientRT})

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

func TestList_dropsItem_whenPubDateInvalid(t *testing.T) {
	// @given pubDate が壊れた item と正常 item を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(
		rssItemFixture{
			title:   "壊れた時刻",
			link:    "https://techcrunch.com/bad/",
			pubDate: "not-a-timestamp",
			guid:    "https://techcrunch.com/?p=bad",
		},
		rssItemFixture{
			title:   "正常",
			link:    "https://techcrunch.com/ok/",
			pubDate: formatRFC1123Z(since.Add(time.Hour)),
			guid:    "https://techcrunch.com/?p=ok",
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
	if got[0].Summary != "正常" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "正常")
	}
}

func TestList_discardsContentEncoded_whenPresent(t *testing.T) {
	// @given 空でない content:encoded を持つ item double（写像では捨てる）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rssXML(rssItemFixture{
		title:          "encoded 捨て",
		link:           "https://techcrunch.com/encoded/",
		description:    "説明のみ",
		pubDate:        formatRFC1123Z(since.Add(time.Hour)),
		guid:           "https://techcrunch.com/?p=enc",
		contentEncoded: "<p>本文フル HTML は捨てる</p>",
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then Detail.Text は description のみ
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].Detail.Text != "説明のみ" {
		t.Fatalf("Detail.Text = %q, want description only", got[0].Detail.Text)
	}
	if strings.Contains(got[0].Detail.Text, "フル HTML") {
		t.Fatalf("content:encoded leaked into Detail.Text: %q", got[0].Detail.Text)
	}
}

func assertTechCrunchInfraError(t *testing.T, err error) {
	t.Helper()
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "techcrunch:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "techcrunch:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
}

// sequenceRoundTripper は request ごとに任意の応答を返す最小 Transport。
type sequenceRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (rt *sequenceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.handler(req)
}
