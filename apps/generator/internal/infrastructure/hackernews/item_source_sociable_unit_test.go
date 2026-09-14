package hackernews_test

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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
)

// stubClientResponse は RoundTrip 1 回分の応答または error を表す。
type stubClientResponse struct {
	Status int
	Body   string
	Err    error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
// topstories.json と item/<id>.json を URL path suffix で出し分け、
// 各 path の呼び出し回数を calls へ記録する。
type stubRoundTripper struct {
	byPath  map[string]stubClientResponse
	fixed   *stubClientResponse
	calls   map[string]int
	callLog []string
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
	rt.callLog = append(rt.callLog, path)

	res, ok := rt.byPath[path]
	if !ok {
		if rt.fixed != nil {
			res = *rt.fixed
		} else {
			return nil, fmt.Errorf("stubRoundTripper: no response configured for path %q", path)
		}
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

// setTopStories は topstories.json の応答を id 列から組む。
func (rt *stubRoundTripper) setTopStories(ids ...int) {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("%d", id))
	}
	rt.byPath["/v0/topstories.json"] = stubClientResponse{Body: "[" + strings.Join(parts, ",") + "]"}
}

// setItem は item/<id>.json の応答を raw JSON body から組む。
func (rt *stubRoundTripper) setItem(id int, body string) {
	rt.byPath[fmt.Sprintf("/v0/item/%d.json", id)] = stubClientResponse{Body: body}
}

// setItemResponse は item/<id>.json に status / error を含む応答を仕込む。
func (rt *stubRoundTripper) setItemResponse(id int, res stubClientResponse) {
	rt.byPath[fmt.Sprintf("/v0/item/%d.json", id)] = res
}

func (rt *stubRoundTripper) itemCalls(id int) int {
	return rt.calls[fmt.Sprintf("/v0/item/%d.json", id)]
}

func newStubListItemSource(rt *stubRoundTripper) *hackernews.ListItemSource {
	return hackernews.NewListItemSource(&http.Client{Transport: rt})
}

// storyJSON は type=="story" の item JSON を組む helper。
func storyJSON(id int, unix int64, title, text, url string, kids ...int) string {
	kidParts := make([]string, 0, len(kids))
	for _, k := range kids {
		kidParts = append(kidParts, fmt.Sprintf("%d", k))
	}
	return fmt.Sprintf(
		`{"id":%d,"type":"story","by":"user%d","time":%d,"title":%q,"text":%q,"url":%q,"kids":[%s]}`,
		id, id, unix, title, text, url, strings.Join(kidParts, ","),
	)
}

// commentJSON は type=="comment" の item JSON を組む helper。
func commentJSON(id int, unix int64, by, text string) string {
	return fmt.Sprintf(`{"id":%d,"type":"comment","by":%q,"time":%d,"text":%q}`, id, by, unix, text)
}

func TestList_mapsTopStoryToSourceItem_whenStoryInWindow(t *testing.T) {
	// @given topstories.json→item/<id>.json の double。story 1 件（window 内、kids なし）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	storyUnix := since.Add(3 * time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(101)
	rt.setItem(101, storyJSON(101, storyUnix, "タイトル本文", "本文テキスト", "https://example.com/article"))
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
	wantOccurredAt := time.Unix(storyUnix, 0).UTC()
	if got[0].SourceID != hackernews.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, hackernews.SourceID)
	}
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
	if got[0].Summary != "タイトル本文" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "タイトル本文")
	}
	if got[0].Detail.Text != "本文テキスト" {
		t.Fatalf("Detail.Text = %q, want %q", got[0].Detail.Text, "本文テキスト")
	}
	if len(got[0].Detail.Links) != 1 || got[0].Detail.Links[0] != "https://example.com/article" {
		t.Fatalf("Detail.Links = %#v, want article URL", got[0].Detail.Links)
	}
	if got[0].Discourse.Text != "" {
		t.Fatalf("Discourse.Text = %q, want empty", got[0].Discourse.Text)
	}
	wantPermalink := "https://news.ycombinator.com/item?id=101"
	if len(got[0].Discourse.Links) != 1 || got[0].Discourse.Links[0] != wantPermalink {
		t.Fatalf("Discourse.Links = %#v, want %q", got[0].Discourse.Links, wantPermalink)
	}
	if !strings.Contains(got[0].Meta, "item_id: 101") {
		t.Fatalf("Meta = %q, want item_id", got[0].Meta)
	}
}

func TestList_filtersToTypeStory_whenJobOrPollPresent(t *testing.T) {
	// @given job / poll / story の id を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(1, 2, 3)
	rt.setItem(1, fmt.Sprintf(`{"id":1,"type":"job","by":"u1","time":%d,"title":"求人"}`, unix))
	rt.setItem(2, fmt.Sprintf(`{"id":2,"type":"poll","by":"u2","time":%d,"title":"投票"}`, unix))
	rt.setItem(3, storyJSON(3, unix, "記事", "", ""))
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
	if got[0].Summary != "記事" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "記事")
	}
}

func TestList_excludesDeletedOrDeadItems(t *testing.T) {
	// @given deleted:true / dead:true / 正常 story を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(10, 11, 12)
	rt.setItem(10, fmt.Sprintf(`{"id":10,"type":"story","by":"u","time":%d,"title":"消し","deleted":true}`, unix))
	rt.setItem(11, fmt.Sprintf(`{"id":11,"type":"story","by":"u","time":%d,"title":"死","dead":true}`, unix))
	rt.setItem(12, storyJSON(12, unix, "生存", "", ""))
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
	if got[0].Summary != "生存" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "生存")
	}
}

func TestList_excludesItemsOlderThanSince_atBoundary(t *testing.T) {
	// @given time==since の story と time==since-1s の story を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setTopStories(20, 21)
	rt.setItem(20, storyJSON(20, since.Unix(), "境界ちょうど", "", ""))
	rt.setItem(21, storyJSON(21, since.Add(-time.Second).Unix(), "境界の外", "", ""))
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

func TestList_fetchesTopLevelCommentsUpToMaxCommentsPerStory(t *testing.T) {
	// @given kids を MaxCommentsPerStory 超で持つ story の double。全 kid comment を仕込む
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	kidCount := hackernews.MaxCommentsPerStory + 5
	kids := make([]int, 0, kidCount)
	for i := 0; i < kidCount; i++ {
		kids = append(kids, 1000+i)
	}
	rt := newStubRoundTripper()
	rt.setTopStories(30)
	rt.setItem(30, storyJSON(30, unix, "コメント上限テスト", "", "", kids...))
	for _, k := range kids {
		rt.setItem(k, commentJSON(k, unix, "commenter", fmt.Sprintf("コメント%d", k)))
	}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と comment fetch 回数
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	fetched := 0
	for _, k := range kids {
		fetched += rt.itemCalls(k)
	}
	if fetched != hackernews.MaxCommentsPerStory {
		t.Fatalf("fetched comment count = %d, want %d", fetched, hackernews.MaxCommentsPerStory)
	}
	// 上限を超えた kid は fetch されない。
	if rt.itemCalls(kids[hackernews.MaxCommentsPerStory]) != 0 {
		t.Fatalf("over-limit kid was fetched (calls=%d)", rt.itemCalls(kids[hackernews.MaxCommentsPerStory]))
	}
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	// @given 全 story が since より古い double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setTopStories(40, 41)
	rt.setItem(40, storyJSON(40, since.Add(-time.Hour).Unix(), "古い1", "", ""))
	rt.setItem(41, storyJSON(41, since.Add(-24*time.Hour).Unix(), "古い2", "", ""))
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

func TestList_returnsInfrastructureError_whenClientNilOrNon200OrInvalidJSON(t *testing.T) {
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("client nil", func(t *testing.T) {
		// @given client を持たない ListItemSource
		source := hackernews.NewListItemSource(nil)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *adaptererror.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
		}
		if !strings.HasPrefix(infra.Error(), "hackernews:") {
			t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "hackernews:")
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
	})

	t.Run("non-200 topstories", func(t *testing.T) {
		// @given topstories.json が 500 を返し、再試行でも 500 を返す double
		rt := newStubRoundTripper()
		rt.byPath["/v0/topstories.json"] = stubClientResponse{Status: http.StatusInternalServerError, Body: "boom"}
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *adaptererror.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
		}
		if !strings.HasPrefix(infra.Error(), "hackernews:") {
			t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "hackernews:")
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
	})

	t.Run("invalid JSON topstories", func(t *testing.T) {
		// @given topstories.json が壊れた JSON を返す double
		rt := newStubRoundTripper()
		rt.byPath["/v0/topstories.json"] = stubClientResponse{Body: "not-json"}
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		var infra *adaptererror.Error
		if !errors.As(err, &infra) {
			t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
		}
		if !strings.HasPrefix(infra.Error(), "hackernews:") {
			t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "hackernews:")
		}
		if errors.Unwrap(infra) == nil {
			t.Fatal("Unwrap() = nil, want non-nil")
		}
	})
}

func TestList_putsCommentTextIntoTextLine_whenStoryTextEmpty(t *testing.T) {
	// @given text 空・kids ありの story と、その kid comment の double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(50)
	rt.setItem(50, storyJSON(50, unix, "外部リンク記事", "", "https://example.com/x", 501, 502))
	rt.setItem(501, commentJSON(501, unix, "alice", "一つ目のコメント"))
	rt.setItem(502, commentJSON(502, unix, "bob", "二つ目のコメント"))
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
	if got[0].Summary != "外部リンク記事" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "外部リンク記事")
	}
	if got[0].Detail.Text != "" {
		t.Fatalf("Detail.Text = %q, want empty", got[0].Detail.Text)
	}
	if len(got[0].Detail.Links) != 1 || got[0].Detail.Links[0] != "https://example.com/x" {
		t.Fatalf("Detail.Links = %#v, want external URL", got[0].Detail.Links)
	}
	wantDiscourse := "一つ目のコメント\n二つ目のコメント"
	if got[0].Discourse.Text != wantDiscourse {
		t.Fatalf("Discourse.Text = %q, want %q", got[0].Discourse.Text, wantDiscourse)
	}
	if len(got[0].Discourse.Links) != 1 || got[0].Discourse.Links[0] != "https://news.ycombinator.com/item?id=50" {
		t.Fatalf("Discourse.Links = %#v, want permalink", got[0].Discourse.Links)
	}
}

func TestList_omitsLinksLine_whenStoryURLAbsent(t *testing.T) {
	// @given url を持つ story と持たない story の double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(60, 61)
	rt.setItem(60, storyJSON(60, unix, "URLあり", "本文", "https://example.com/withurl"))
	rt.setItem(61, storyJSON(61, unix, "URLなし", "本文", ""))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	summaries := map[string]bool{}
	var withURL, withoutURL models.SourceItem
	for _, item := range got {
		summaries[item.Summary] = true
		switch item.Summary {
		case "URLあり":
			withURL = item
		case "URLなし":
			withoutURL = item
		}
	}
	if !summaries["URLあり"] || !summaries["URLなし"] {
		t.Fatalf("Summaries = %v, want URLあり and URLなし", summaries)
	}
	if len(withURL.Detail.Links) != 1 || withURL.Detail.Links[0] != "https://example.com/withurl" {
		t.Fatalf("URLあり Detail.Links = %#v", withURL.Detail.Links)
	}
	if len(withoutURL.Detail.Links) != 0 {
		t.Fatalf("URLなし Detail.Links = %#v, want empty", withoutURL.Detail.Links)
	}
}

func TestList_dropsFailedCommentButKeepsRest_whenOneCommentFetchFails(t *testing.T) {
	// @given kid が 3 件で、うち 1 件の item fetch が失敗する double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(70, 80)
	rt.setItem(70, storyJSON(70, unix, "コメント一部失敗", "", "", 701, 702, 703))
	rt.setItem(701, commentJSON(701, unix, "a", "コメント1"))
	rt.setItemResponse(702, stubClientResponse{Status: http.StatusInternalServerError, Body: "boom"})
	rt.setItem(703, commentJSON(703, unix, "c", "コメント3"))
	rt.setItem(80, storyJSON(80, unix, "別story", "別本文", ""))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2 (%+v)", len(got), got)
	}
	summaries := map[string]bool{}
	for _, item := range got {
		summaries[item.Summary] = true
	}
	if !summaries["コメント一部失敗"] || !summaries["別story"] {
		t.Fatalf("Summaries = %v, want both stories", summaries)
	}
}

func TestList_failsEntirely_whenTopStoriesFetchFails(t *testing.T) {
	// @given topstories.json の client.Do が失敗し、再試行でも失敗する double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.byPath["/v0/topstories.json"] = stubClientResponse{Err: errors.New("connection reset")}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
}

func TestList_retriesOnceOnTransientError_whenSecondAttemptSucceeds(t *testing.T) {
	// @given topstories.json が 1 回目 5xx、2 回目成功を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	topCalls := 0
	// path 別 stub では 1 回ごとの出し分けができないため、専用の Transport を組む。
	transientRT := &sequenceRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			if strings.HasSuffix(req.URL.Path, "/topstories.json") {
				topCalls++
				if topCalls == 1 {
					return &http.Response{
						StatusCode: http.StatusBadGateway,
						Body:       io.NopCloser(strings.NewReader("boom")),
						Header:     make(http.Header),
						Request:    req,
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("[90]")),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(storyJSON(90, unix, "再試行成功", "本文", ""))),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}
	source := hackernews.NewListItemSource(&http.Client{Transport: transientRT})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と topstories fetch 回数
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if topCalls != 2 {
		t.Fatalf("topstories fetch attempts = %d, want 2 (retry once)", topCalls)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
}

func TestList_dropsStory_whenItemJSONIsBroken(t *testing.T) {
	// @given topstories は 2 件。1 件目の item/<id>.json が壊れた JSON を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(110, 111)
	rt.setItem(110, "not-json")
	rt.setItem(111, storyJSON(111, unix, "正常", "本文", ""))
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

func TestList_normalizesHTMLInStoryText_whenTagsAndEntitiesPresent(t *testing.T) {
	// @given story text に <p> 段落・他タグ・HTML entity を含む double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(120)
	rt.setItem(120, storyJSON(
		120, unix, "正規化テスト",
		"最初の段落<p>次の段落 <a href=\\\"https://e.example\\\">link</a> &amp; &lt;tag&gt;",
		"",
	))
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
	if got[0].Summary != "正規化テスト" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "正規化テスト")
	}
}

func TestList_skipsCommentWithEmptyBody_whenCommentTextBlank(t *testing.T) {
	// @given kids 2 件で、片方の comment text が空の double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	rt := newStubRoundTripper()
	rt.setTopStories(130)
	rt.setItem(130, storyJSON(130, unix, "空コメント混在", "", "", 1301, 1302))
	rt.setItem(1301, commentJSON(1301, unix, "a", ""))
	rt.setItem(1302, commentJSON(1302, unix, "b", "中身あり"))
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
	if got[0].Summary != "空コメント混在" {
		t.Fatalf("Summary = %q, want %q", got[0].Summary, "空コメント混在")
	}
}

func TestList_stopsScanningAfterCollectingMaxStoriesScanned_whenEnoughStoriesInWindow(t *testing.T) {
	// @given topstories が MaxStoriesScanned+3 件。全 id を window 内 story として仕込む
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unix := since.Add(time.Hour).Unix()
	total := hackernews.MaxStoriesScanned + 3
	ids := make([]int, 0, total)
	for i := 0; i < total; i++ {
		ids = append(ids, 2000+i)
	}
	rt := newStubRoundTripper()
	rt.setTopStories(ids...)
	for _, id := range ids {
		rt.setItem(id, storyJSON(id, unix, "story", "本文", ""))
	}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と fetch 回数
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != hackernews.MaxStoriesScanned {
		t.Fatalf("len(got) = %d, want %d", len(got), hackernews.MaxStoriesScanned)
	}
	// 上限件数に達した後、次の id は fetch しない。
	if rt.itemCalls(ids[hackernews.MaxStoriesScanned]) != 0 {
		t.Fatalf("story fetched after limit reached (calls=%d)", rt.itemCalls(ids[hackernews.MaxStoriesScanned]))
	}
}

func TestList_collectsWindowStoriesBeyondFirstMaxStoriesScanned_whenTopStoriesExceedsLimit(t *testing.T) {
	// @given topstories 先頭 5 件は window 外、その後ろに window 内 story を MaxStoriesScanned+2 件仕込む
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	oldUnix := since.Add(-48 * time.Hour).Unix()
	freshUnix := since.Add(time.Hour).Unix()

	oldIDs := []int{3000, 3001, 3002, 3003, 3004}
	freshCount := hackernews.MaxStoriesScanned + 2
	freshIDs := make([]int, 0, freshCount)
	for i := 0; i < freshCount; i++ {
		freshIDs = append(freshIDs, 3100+i)
	}
	allIDs := append(append([]int{}, oldIDs...), freshIDs...)

	rt := newStubRoundTripper()
	rt.setTopStories(allIDs...)
	for _, id := range oldIDs {
		rt.setItem(id, storyJSON(id, oldUnix, "古い", "本文", ""))
	}
	for _, id := range freshIDs {
		rt.setItem(id, storyJSON(id, freshUnix, "新しい", "本文", ""))
	}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != hackernews.MaxStoriesScanned {
		t.Fatalf("len(got) = %d, want %d", len(got), hackernews.MaxStoriesScanned)
	}
	// 先頭 30 件目より後ろ（先頭に window 外が混じっているぶんズレる）の window 内 story も拾う。
	if got[hackernews.MaxStoriesScanned-1].Summary != "新しい" {
		t.Fatalf("last Summary = %q, want %q", got[hackernews.MaxStoriesScanned-1].Summary, "新しい")
	}
}

func TestList_scansEntireIDListForWindowStories_whenOldStoriesPrecedeFreshOnes(t *testing.T) {
	// @given topstories 先頭に window 外 story を 30 件、その後ろに window 内 story を 3 件仕込む
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	oldUnix := since.Add(-72 * time.Hour).Unix()
	freshUnix := since.Add(2 * time.Hour).Unix()

	oldIDs := make([]int, 0, 30)
	for i := 0; i < 30; i++ {
		oldIDs = append(oldIDs, 4000+i)
	}
	freshIDs := []int{4100, 4101, 4102}
	allIDs := append(append([]int{}, oldIDs...), freshIDs...)

	rt := newStubRoundTripper()
	rt.setTopStories(allIDs...)
	for _, id := range oldIDs {
		rt.setItem(id, storyJSON(id, oldUnix, "古い", "本文", ""))
	}
	for _, id := range freshIDs {
		rt.setItem(id, storyJSON(id, freshUnix, "新しい", "本文", ""))
	}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != len(freshIDs) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(freshIDs))
	}
	for i := range freshIDs {
		if got[i].Summary != "新しい" {
			t.Fatalf("got[%d].Summary = %q, want %q", i, got[i].Summary, "新しい")
		}
	}
}

func TestList_failsAfterRetry_whenSecondTopStoriesAttemptAlsoFails(t *testing.T) {
	// @given topstories.json が 2 回とも 5xx を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	attempts := 0
	rt := &sequenceRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			attempts++
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader("down")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}
	source := hackernews.NewListItemSource(&http.Client{Transport: rt})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と再試行回数
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2 (retry once)", attempts)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
}

func TestList_returnsInfrastructureError_whenResponseBodyReadFails(t *testing.T) {
	// @given topstories.json の応答 body が読み取り中に必ず error を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := &sequenceRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(errReader{}),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		},
	}
	source := hackernews.NewListItemSource(&http.Client{Transport: rt})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "hackernews:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "hackernews:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
}

// errReader は Read が常に error を返す io.Reader。io.ReadAll 失敗経路の検証に使う。
type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("body read failed")
}

// sequenceRoundTripper は request ごとに任意の応答を返す最小 Transport。
type sequenceRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (rt *sequenceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.handler(req)
}
