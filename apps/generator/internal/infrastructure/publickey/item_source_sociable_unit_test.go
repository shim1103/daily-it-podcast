package publickey_test

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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/publickey"
)

const feedPath = "/atom.xml"

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

func (rt *stubRoundTripper) feedCalls() int {
	return rt.calls[feedPath]
}

func newStubListItemSource(rt *stubRoundTripper) *publickey.ListItemSource {
	return publickey.NewListItemSource(&http.Client{Transport: rt})
}

// atomEntryFixture は Atom entry XML を組むための入力。
type atomEntryFixture struct {
	title     string
	id        string
	published string
	updated   string
	summary   string
	content   string
	href      string
	author    string
}

func atomXML(entries ...atomEntryFixture) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<feed xmlns="http://www.w3.org/2005/Atom">` + "\n")
	b.WriteString("<title>Publickey</title>\n")
	for _, e := range entries {
		b.WriteString("<entry>\n")
		if e.title != "" {
			fmt.Fprintf(&b, "<title>%s</title>\n", e.title)
		}
		if e.href != "" {
			fmt.Fprintf(&b, `<link rel="alternate" type="text/html" href="%s" />`+"\n", e.href)
		}
		b.WriteString(`<link rel="self" type="application/atom+xml" href="https://www.publickey1.jp/atom.xml" />` + "\n")
		if e.id != "" {
			fmt.Fprintf(&b, "<id>%s</id>\n", e.id)
		}
		if e.published != "" {
			fmt.Fprintf(&b, "<published>%s</published>\n", e.published)
		}
		if e.updated != "" {
			fmt.Fprintf(&b, "<updated>%s</updated>\n", e.updated)
		}
		if e.summary != "" {
			fmt.Fprintf(&b, "<summary>%s</summary>\n", e.summary)
		}
		if e.author != "" {
			fmt.Fprintf(&b, "<author><name>%s</name></author>\n", e.author)
		}
		if e.content != "" {
			if strings.Contains(e.content, "<") {
				fmt.Fprintf(&b, "<content type=\"html\"><![CDATA[%s]]></content>\n", e.content)
			} else {
				fmt.Fprintf(&b, "<content type=\"html\">%s</content>\n", e.content)
			}
		}
		b.WriteString("</entry>\n")
	}
	b.WriteString("</feed>\n")
	return b.String()
}

func TestList_mapsAtomEntryToSourceItem_whenEntryInWindow(t *testing.T) {
	// @given Atom feed double。window 内 entry 1 件（summary / content / alternate link / author あり）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	published := since.Add(3 * time.Hour)
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(atomEntryFixture{
		title:     "タイトル本文",
		id:        "tag:www.publickey1.jp,2026://2.8701",
		published: published.Format(time.RFC3339),
		updated:   published.Add(time.Minute).Format(time.RFC3339),
		summary:   "要約テキスト",
		content:   "最初の段落<p>次の段落 <a href=\"https://e.example\">link</a> &amp; &lt;tag&gt;",
		href:      "https://www.publickey1.jp/blog/26/example.html",
		author:    "jniino",
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
	if got[0].SourceID != publickey.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, publickey.SourceID)
	}
	wantOccurredAt := published.UTC()
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
	wantSummary := "タイトル本文\n要約テキスト"
	if got[0].Summary != wantSummary {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, wantSummary)
	}
	wantDetailText := "最初の段落\n次の段落 link & <tag>"
	if got[0].Detail.Text != wantDetailText {
		t.Fatalf("Detail.Text = %q, want %q", got[0].Detail.Text, wantDetailText)
	}
	if len(got[0].Detail.Links) != 1 || got[0].Detail.Links[0] != "https://www.publickey1.jp/blog/26/example.html" {
		t.Fatalf("Detail.Links = %#v, want alternate href", got[0].Detail.Links)
	}
	if !got[0].Discourse.Empty() {
		t.Fatalf("Discourse = %+v, want empty", got[0].Discourse)
	}
	if !strings.Contains(got[0].Meta, "item_id: tag:www.publickey1.jp,2026://2.8701") {
		t.Fatalf("Meta = %q, want item_id", got[0].Meta)
	}
	if !strings.Contains(got[0].Meta, "actor_id: jniino") || !strings.Contains(got[0].Meta, "actor_name: jniino") {
		t.Fatalf("Meta = %q, want actor from author", got[0].Meta)
	}
}

func TestList_excludesEntriesOlderThanSince_atBoundary(t *testing.T) {
	// @given published==since の entry と published==since-1s の entry を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(
		atomEntryFixture{
			title:     "境界ちょうど",
			id:        "in",
			published: since.Format(time.RFC3339),
			href:      "https://www.publickey1.jp/in.html",
		},
		atomEntryFixture{
			title:     "境界の外",
			id:        "out",
			published: since.Add(-time.Second).Format(time.RFC3339),
			href:      "https://www.publickey1.jp/out.html",
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

func TestList_stopsAfterCollectingMaxStoriesScanned_whenEnoughEntriesInWindow(t *testing.T) {
	// @given window 内 entry を MaxStoriesScanned+3 件持つ Atom double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	total := publickey.MaxStoriesScanned + 3
	entries := make([]atomEntryFixture, 0, total)
	for i := 0; i < total; i++ {
		entries = append(entries, atomEntryFixture{
			title:     fmt.Sprintf("記事%d", i),
			id:        fmt.Sprintf("id-%d", i),
			published: since.Add(time.Duration(i+1) * time.Minute).Format(time.RFC3339),
			href:      fmt.Sprintf("https://www.publickey1.jp/%d.html", i),
		})
	}
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(entries...))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 結果は MaxStoriesScanned 件で打ち切る
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != publickey.MaxStoriesScanned {
		t.Fatalf("len(got) = %d, want %d", len(got), publickey.MaxStoriesScanned)
	}
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	// @given 全 entry が since より古い double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(
		atomEntryFixture{
			title:     "古い1",
			id:        "old1",
			published: since.Add(-time.Hour).Format(time.RFC3339),
		},
		atomEntryFixture{
			title:     "古い2",
			id:        "old2",
			published: since.Add(-24 * time.Hour).Format(time.RFC3339),
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
		source := publickey.NewListItemSource(nil)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		assertPublickeyInfraError(t, err)
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
		assertPublickeyInfraError(t, err)
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
		assertPublickeyInfraError(t, err)
	})
}

func TestList_retriesOnceOnTransientError_whenSecondAttemptSucceeds(t *testing.T) {
	// @given feed が 1 回目 5xx、2 回目成功を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	published := since.Add(time.Hour).Format(time.RFC3339)
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
			body := atomXML(atomEntryFixture{
				title:     "再試行成功",
				id:        "retry1",
				published: published,
				href:      "https://www.publickey1.jp/retry.html",
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}
	source := publickey.NewListItemSource(&http.Client{Transport: transientRT})

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

func TestList_dropsEntry_whenPublishedAndUpdatedAreInvalid(t *testing.T) {
	// @given 時刻が壊れた entry と正常 entry を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(
		atomEntryFixture{
			title:     "壊れた時刻",
			id:        "bad",
			published: "not-a-timestamp",
			updated:   "also-bad",
			href:      "https://www.publickey1.jp/bad.html",
		},
		atomEntryFixture{
			title:     "正常",
			id:        "ok",
			published: since.Add(time.Hour).Format(time.RFC3339),
			href:      "https://www.publickey1.jp/ok.html",
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

func TestList_usesUpdated_whenPublishedAbsent(t *testing.T) {
	// @given published 無し・updated のみの entry double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	updated := since.Add(2 * time.Hour)
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(atomEntryFixture{
		title:   "updated のみ",
		id:      "upd1",
		updated: updated.Format(time.RFC3339),
		href:    "https://www.publickey1.jp/upd.html",
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
	if !got[0].OccurredAt.Equal(updated.UTC()) {
		t.Fatalf("OccurredAt = %v, want %v", got[0].OccurredAt, updated.UTC())
	}
}

func TestList_selectsAlternateLinkOnly_whenSelfLinkAlsoPresent(t *testing.T) {
	// @given alternate と self の link を持つ entry double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setFeed(atomXML(atomEntryFixture{
		title:     "link 選別",
		id:        "link1",
		published: since.Add(time.Hour).Format(time.RFC3339),
		href:      "https://www.publickey1.jp/article.html",
	}))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then Detail.Links は alternate のみ
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if len(got[0].Detail.Links) != 1 || got[0].Detail.Links[0] != "https://www.publickey1.jp/article.html" {
		t.Fatalf("Detail.Links = %#v, want alternate only", got[0].Detail.Links)
	}
}

func assertPublickeyInfraError(t *testing.T, err error) {
	t.Helper()
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "publickey:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "publickey:")
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
