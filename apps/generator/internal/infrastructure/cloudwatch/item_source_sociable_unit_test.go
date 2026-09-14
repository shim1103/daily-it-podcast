package cloudwatch_test

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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/cloudwatch"
)

const feedPath = "/data/rss/1.0/clw/feed.rdf"

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

func newStubListItemSource(rt *stubRoundTripper) *cloudwatch.ListItemSource {
	return cloudwatch.NewListItemSource(&http.Client{Transport: rt})
}

// rdfItemFixture は RDF item XML を組むための入力。
type rdfItemFixture struct {
	about       string
	title       string
	link        string
	date        string
	description string
	creator     string
	subject     string
}

func rdfXML(items ...rdfItemFixture) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<rdf:RDF xmlns="http://purl.org/rss/1.0/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:dc="http://purl.org/dc/elements/1.1/" xml:lang="ja">` + "\n")
	b.WriteString(` <channel rdf:about="https://cloud.watch.impress.co.jp/data/rss/1.0/clw/feed.rdf">` + "\n")
	b.WriteString("  <title>クラウド Watch</title>\n")
	b.WriteString(" </channel>\n")
	for _, it := range items {
		about := it.about
		if about == "" && it.link != "" {
			about = it.link + "?ref=rss"
		}
		fmt.Fprintf(&b, ` <item rdf:about="%s">`+"\n", about)
		if it.title != "" {
			fmt.Fprintf(&b, "  <title>%s</title>\n", it.title)
		}
		if it.link != "" {
			fmt.Fprintf(&b, "  <link>%s</link>\n", it.link)
		}
		if it.date != "" {
			fmt.Fprintf(&b, "  <dc:date>%s</dc:date>\n", it.date)
		}
		if it.creator != "" {
			fmt.Fprintf(&b, "  <dc:creator>%s</dc:creator>\n", it.creator)
		}
		if it.subject != "" {
			fmt.Fprintf(&b, "  <dc:subject>%s</dc:subject>\n", it.subject)
		}
		if it.description != "" {
			fmt.Fprintf(&b, "  <description><![CDATA[%s]]></description>\n", it.description)
		}
		b.WriteString(" </item>\n")
	}
	b.WriteString("</rdf:RDF>\n")
	return b.String()
}

func TestList_mapsRDFItemToSourceItem_whenItemInWindow(t *testing.T) {
	// @given RDF feed double。window 内 item 1 件（link は ?ref 無し、rdf:about は ?ref=rss、dc:subject / creator あり）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	articleURL := "https://cloud.watch.impress.co.jp/docs/news/2140188.html"
	occurredAt := time.Date(2026, 9, 1, 10, 0, 0, 0, time.FixedZone("JST", 9*3600))
	rt := newStubRoundTripper()
	rt.setFeed(rdfXML(rdfItemFixture{
		about:       articleURL + "?ref=rss",
		title:       "タイトル本文",
		link:        articleURL,
		date:        occurredAt.Format(time.RFC3339),
		description: "説明テキスト",
		creator:     "石井 一志",
		subject:     "クラウド",
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
	if got[0].SourceID != cloudwatch.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, cloudwatch.SourceID)
	}
	wantOccurredAt := occurredAt.UTC()
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
	if got[0].Summary != "タイトル本文" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "タイトル本文")
	}
	if got[0].Detail.Text != "説明テキスト" {
		t.Fatalf("Detail.Text = %q, want description", got[0].Detail.Text)
	}
	if len(got[0].Detail.Links) != 1 || got[0].Detail.Links[0] != articleURL {
		t.Fatalf("Detail.Links = %#v, want link without ?ref", got[0].Detail.Links)
	}
	if strings.Contains(got[0].Detail.Links[0], "ref=rss") {
		t.Fatalf("Detail.Links used rdf:about: %#v", got[0].Detail.Links)
	}
	if !got[0].Discourse.Empty() {
		t.Fatalf("Discourse = %+v, want empty", got[0].Discourse)
	}
	if got[0].Meta != "" {
		t.Fatalf("Meta = %q, want empty (作者無し契約)", got[0].Meta)
	}
	if strings.Contains(got[0].Summary, "クラウド") || strings.Contains(got[0].Detail.Text, "クラウド") {
		t.Fatalf("dc:subject leaked: Summary=%q Detail=%q", got[0].Summary, got[0].Detail.Text)
	}
}

func TestList_excludesItemsOlderThanSince_atBoundary(t *testing.T) {
	// @given dc:date==since の item と dc:date==since-1s の item を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rdfXML(
		rdfItemFixture{
			title: "境界ちょうど",
			link:  "https://cloud.watch.impress.co.jp/docs/news/in.html",
			date:  since.Format(time.RFC3339),
		},
		rdfItemFixture{
			title: "境界の外",
			link:  "https://cloud.watch.impress.co.jp/docs/news/out.html",
			date:  since.Add(-time.Second).Format(time.RFC3339),
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

func TestList_stopsAfterCollectingMaxStoriesScanned_whenEnoughItemsInWindow(t *testing.T) {
	// @given window 内 item を MaxStoriesScanned+3 件持つ RDF double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	total := cloudwatch.MaxStoriesScanned + 3
	items := make([]rdfItemFixture, 0, total)
	for i := 0; i < total; i++ {
		link := fmt.Sprintf("https://cloud.watch.impress.co.jp/docs/news/%d.html", i)
		items = append(items, rdfItemFixture{
			title:       fmt.Sprintf("記事%d", i),
			link:        link,
			date:        since.Add(time.Duration(i+1) * time.Minute).Format(time.RFC3339),
			description: "説明",
		})
	}
	rt := newStubRoundTripper()
	rt.setFeed(rdfXML(items...))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 結果は MaxStoriesScanned 件で打ち切る
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != cloudwatch.MaxStoriesScanned {
		t.Fatalf("len(got) = %d, want %d", len(got), cloudwatch.MaxStoriesScanned)
	}
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	// @given 全 item が since より古い double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rdfXML(
		rdfItemFixture{
			title: "古い1",
			link:  "https://cloud.watch.impress.co.jp/docs/news/old1.html",
			date:  since.Add(-time.Hour).Format(time.RFC3339),
		},
		rdfItemFixture{
			title: "古い2",
			link:  "https://cloud.watch.impress.co.jp/docs/news/old2.html",
			date:  since.Add(-24 * time.Hour).Format(time.RFC3339),
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
		source := cloudwatch.NewListItemSource(nil)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		assertCloudWatchInfraError(t, err)
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
		assertCloudWatchInfraError(t, err)
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
		assertCloudWatchInfraError(t, err)
	})
}

func TestList_retriesOnceOnTransientError_whenSecondAttemptSucceeds(t *testing.T) {
	// @given feed が 1 回目 5xx、2 回目成功を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	date := since.Add(time.Hour).Format(time.RFC3339)
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
			body := rdfXML(rdfItemFixture{
				title: "再試行成功",
				link:  "https://cloud.watch.impress.co.jp/docs/news/retry.html",
				date:  date,
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}
	source := cloudwatch.NewListItemSource(&http.Client{Transport: transientRT})

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

func TestList_dropsItem_whenDateInvalid(t *testing.T) {
	// @given dc:date が壊れた item と正常 item を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rdfXML(
		rdfItemFixture{
			title: "壊れた時刻",
			link:  "https://cloud.watch.impress.co.jp/docs/news/bad.html",
			date:  "not-a-timestamp",
		},
		rdfItemFixture{
			title: "正常",
			link:  "https://cloud.watch.impress.co.jp/docs/news/ok.html",
			date:  since.Add(time.Hour).Format(time.RFC3339),
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

func TestList_parsesDateWithTimezoneOffset_toUTC(t *testing.T) {
	// @given dc:date が +09:00 offset 付きの double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	date := "2026-09-01T12:00:00+09:00"
	wantOccurredAt := time.Date(2026, 9, 1, 3, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(rdfXML(rdfItemFixture{
		title: "TZ テスト",
		link:  "https://cloud.watch.impress.co.jp/docs/news/tz.html",
		date:  date,
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と OccurredAt
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
}

func assertCloudWatchInfraError(t *testing.T, err error) {
	t.Helper()
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "cloudwatch:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "cloudwatch:")
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
