package lobsters_test

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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/lobsters"
)

// stubClientResponse は RoundTrip 1 回分の応答または error を表す。
type stubClientResponse struct {
	Status int
	Body   string
	Err    error
}

// stubRoundTripper は http.RoundTripper を境界 I/O なしで満たす直接 Stub。
// hottest.json と /s/<short_id>.json を URL path suffix で出し分け、
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

// setHottest は hottest.json の応答を summary 列から組む。
func (rt *stubRoundTripper) setHottest(entries ...hottestEntry) {
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, fmt.Sprintf(`{"short_id":%q,"created_at":%q}`, e.ShortID, e.CreatedAt))
	}
	rt.byPath["/hottest.json"] = stubClientResponse{Body: "[" + strings.Join(parts, ",") + "]"}
}

// setStory は /s/<short_id>.json の応答を raw JSON body から組む。
func (rt *stubRoundTripper) setStory(shortID, body string) {
	rt.byPath[fmt.Sprintf("/s/%s.json", shortID)] = stubClientResponse{Body: body}
}

// setStoryResponse は /s/<short_id>.json に status / error を含む応答を仕込む。
func (rt *stubRoundTripper) setStoryResponse(shortID string, res stubClientResponse) {
	rt.byPath[fmt.Sprintf("/s/%s.json", shortID)] = res
}

func (rt *stubRoundTripper) storyCalls(shortID string) int {
	return rt.calls[fmt.Sprintf("/s/%s.json", shortID)]
}

func newStubListItemSource(rt *stubRoundTripper) *lobsters.ListItemSource {
	return lobsters.NewListItemSource(&http.Client{Transport: rt})
}

type hottestEntry struct {
	ShortID   string
	CreatedAt string
}

// storyJSON は story 詳細 JSON を組む helper。
func storyJSON(shortID, createdAt, submitter, title, descriptionPlain, url, shortIDURL, commentsURL string, comments ...string) string {
	commentParts := make([]string, 0, len(comments))
	for _, c := range comments {
		commentParts = append(commentParts, c)
	}
	commentsField := "[]"
	if len(commentParts) > 0 {
		commentsField = "[" + strings.Join(commentParts, ",") + "]"
	}
	return fmt.Sprintf(
		`{"short_id":%q,"submitter_user":%q,"title":%q,"description_plain":%q,"url":%q,"short_id_url":%q,"comments_url":%q,"created_at":%q,"comments":%s}`,
		shortID, submitter, title, descriptionPlain, url, shortIDURL, commentsURL, createdAt, commentsField,
	)
}

// commentJSON は comments[] の 1 要素 JSON を組む helper。
func commentJSON(plain string, deleted, moderated bool) string {
	return fmt.Sprintf(`{"comment_plain":%q,"is_deleted":%t,"is_moderated":%t}`, plain, deleted, moderated)
}

func TestList_mapsHottestStoryToSourceItem_whenStoryInWindow(t *testing.T) {
	// @given hottest.json→/s/<short_id>.json の double。story 1 件（window 内、comments なし）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(3 * time.Hour).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(hottestEntry{ShortID: "abc123", CreatedAt: createdAt})
	rt.setStory("abc123", storyJSON(
		"abc123", createdAt, "alice", "タイトル本文", "説明テキスト",
		"https://example.com/article",
		"https://lobste.rs/s/abc123",
		"https://lobste.rs/s/abc123/comments",
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
	wantOccurredAt, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		t.Fatalf("parse createdAt: %v", err)
	}
	wantOccurredAt = wantOccurredAt.UTC()
	if got[0].SourceID != lobsters.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, lobsters.SourceID)
	}
	if !got[0].OccurredAt.Equal(wantOccurredAt) || got[0].OccurredAt.Location() != time.UTC {
		t.Fatalf("OccurredAt = %v, want %v (UTC)", got[0].OccurredAt, wantOccurredAt)
	}
	wantContext := strings.Join([]string{
		"item_id: abc123",
		"actor_id: alice",
		"actor_name: alice",
		"title: タイトル本文",
		"text: 説明テキスト",
		"permalink: https://lobste.rs/s/abc123",
		"links: https://example.com/article",
	}, "\n")
	if got[0].Context != wantContext {
		t.Fatalf("Context =\n%q\nwant\n%q", got[0].Context, wantContext)
	}
}

func TestList_excludesStoriesOlderThanSince_atBoundary(t *testing.T) {
	// @given created_at==since の story と created_at==since-1s の story を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	atBoundary := since.Format(time.RFC3339)
	beforeBoundary := since.Add(-time.Second).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(
		hottestEntry{ShortID: "inwin", CreatedAt: atBoundary},
		hottestEntry{ShortID: "outwin", CreatedAt: beforeBoundary},
	)
	rt.setStory("inwin", storyJSON("inwin", atBoundary, "u", "境界ちょうど", "", "", "https://lobste.rs/s/inwin", ""))
	rt.setStory("outwin", storyJSON("outwin", beforeBoundary, "u", "境界の外", "", "", "https://lobste.rs/s/outwin", ""))
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
	if !strings.Contains(got[0].Context, "item_id: inwin") {
		t.Fatalf("Context = %q, want story inwin (boundary inclusive)", got[0].Context)
	}
}

func TestList_excludesDeletedOrModeratedComments(t *testing.T) {
	// @given deleted / moderated な comment を混ぜた double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(time.Hour).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(hottestEntry{ShortID: "cmt1", CreatedAt: createdAt})
	rt.setStory("cmt1", storyJSON(
		"cmt1", createdAt, "u", "コメント除外", "", "",
		"https://lobste.rs/s/cmt1", "",
		commentJSON("正常コメント", false, false),
		commentJSON("削除済み", true, false),
		commentJSON("モデレート済み", false, true),
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
	if !strings.Contains(got[0].Context, "正常コメント") {
		t.Fatalf("Context = %q, want 正常コメント", got[0].Context)
	}
	if strings.Contains(got[0].Context, "削除済み") || strings.Contains(got[0].Context, "モデレート済み") {
		t.Fatalf("Context = %q, want no deleted/moderated comment text", got[0].Context)
	}
}

func TestList_usesCommentPlainForCommentBody(t *testing.T) {
	// @given comment_plain と HTML comment field を混ぜ、MaxCommentsPerStory 超の comment を仕込む double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(time.Hour).Format(time.RFC3339)
	commentCount := lobsters.MaxCommentsPerStory + 3
	comments := make([]string, 0, commentCount)
	for i := 0; i < commentCount; i++ {
		plain := fmt.Sprintf("plain-%d", i)
		comments = append(comments, fmt.Sprintf(
			`{"comment_plain":%q,"comment":%q,"is_deleted":false,"is_moderated":false}`,
			plain, fmt.Sprintf("<p>html-%d</p>", i),
		))
	}
	rt := newStubRoundTripper()
	rt.setHottest(hottestEntry{ShortID: "plain1", CreatedAt: createdAt})
	rt.setStory("plain1", storyJSON(
		"plain1", createdAt, "u", "plain テスト", "", "",
		"https://lobste.rs/s/plain1", "",
		comments...,
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
	ctx := got[0].Context
	for i := 0; i < lobsters.MaxCommentsPerStory; i++ {
		if !strings.Contains(ctx, fmt.Sprintf("plain-%d", i)) {
			t.Fatalf("Context = %q, want plain-%d", ctx, i)
		}
	}
	if strings.Contains(ctx, "html-") {
		t.Fatalf("Context = %q, want no HTML comment field", ctx)
	}
	for i := lobsters.MaxCommentsPerStory; i < commentCount; i++ {
		if strings.Contains(ctx, fmt.Sprintf("plain-%d", i)) {
			t.Fatalf("Context = %q, want no plain-%d (over MaxCommentsPerStory)", ctx, i)
		}
	}
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	// @given 全 story が since より古い double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	old1 := since.Add(-time.Hour).Format(time.RFC3339)
	old2 := since.Add(-24 * time.Hour).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(
		hottestEntry{ShortID: "old1", CreatedAt: old1},
		hottestEntry{ShortID: "old2", CreatedAt: old2},
	)
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
		source := lobsters.NewListItemSource(nil)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		assertLobstersInfraError(t, err)
	})

	t.Run("non-200 hottest", func(t *testing.T) {
		// @given hottest.json が 500 を返し、再試行でも 500 を返す double
		rt := newStubRoundTripper()
		rt.byPath["/hottest.json"] = stubClientResponse{Status: http.StatusInternalServerError, Body: "boom"}
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		assertLobstersInfraError(t, err)
	})

	t.Run("invalid JSON hottest", func(t *testing.T) {
		// @given hottest.json が壊れた JSON を返す double
		rt := newStubRoundTripper()
		rt.byPath["/hottest.json"] = stubClientResponse{Body: "not-json"}
		source := newStubListItemSource(rt)

		// @when
		got, err := source.List(context.Background(), since)

		// @then 戻り値と error
		if got != nil {
			t.Fatalf("got = %+v, want nil", got)
		}
		assertLobstersInfraError(t, err)
	})
}

func TestList_dropsFailedStoryButKeepsRest_whenOneStoryDetailFetchFails(t *testing.T) {
	// @given hottest に 2 件。1 件目の /s/<short_id>.json が失敗する double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(time.Hour).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(
		hottestEntry{ShortID: "fail1", CreatedAt: createdAt},
		hottestEntry{ShortID: "ok1", CreatedAt: createdAt},
	)
	rt.setStoryResponse("fail1", stubClientResponse{Status: http.StatusInternalServerError, Body: "boom"})
	rt.setStory("ok1", storyJSON("ok1", createdAt, "u", "正常 story", "本文", "", "https://lobste.rs/s/ok1", ""))
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
	if !strings.Contains(got[0].Context, "item_id: ok1") {
		t.Fatalf("Context = %q, want story ok1 only", got[0].Context)
	}
}

func TestList_failsEntirely_whenHottestFetchFails(t *testing.T) {
	// @given hottest.json の client.Do が失敗し、再試行でも失敗する double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.byPath["/hottest.json"] = stubClientResponse{Err: errors.New("connection reset")}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	assertLobstersInfraError(t, err)
}

func assertLobstersInfraError(t *testing.T, err error) {
	t.Helper()
	var infra *lobsters.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *lobsters.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "lobsters:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "lobsters:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
}

func TestList_dropsStory_whenDetailCreatedAtIsInvalid(t *testing.T) {
	// @given hottest summary は window 内だが、story 詳細の created_at が壊れている double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	summaryCreatedAt := since.Add(time.Hour).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(hottestEntry{ShortID: "badts", CreatedAt: summaryCreatedAt})
	rt.setStory("badts", storyJSON("badts", "not-a-timestamp", "u", "壊れた時刻", "", "", "https://lobste.rs/s/badts", ""))
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と error
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0 (%+v)", len(got), got)
	}
}

func TestList_dropsStory_whenStoryJSONIsBroken(t *testing.T) {
	// @given hottest は 2 件。1 件目の /s/<short_id>.json が壊れた JSON を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(time.Hour).Format(time.RFC3339)
	rt := newStubRoundTripper()
	rt.setHottest(
		hottestEntry{ShortID: "broken", CreatedAt: createdAt},
		hottestEntry{ShortID: "ok2", CreatedAt: createdAt},
	)
	rt.setStory("broken", "not-json")
	rt.setStory("ok2", storyJSON("ok2", createdAt, "u", "正常", "本文", "", "https://lobste.rs/s/ok2", ""))
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
	if !strings.Contains(got[0].Context, "item_id: ok2") {
		t.Fatalf("Context = %q, want story ok2 only", got[0].Context)
	}
}

func TestList_stopsScanningAfterCollectingMaxStoriesScanned_whenEnoughStoriesInWindow(t *testing.T) {
	// @given hottest が MaxStoriesScanned+3 件。全 story を window 内として仕込む
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(time.Hour).Format(time.RFC3339)
	total := lobsters.MaxStoriesScanned + 3
	entries := make([]hottestEntry, 0, total)
	for i := 0; i < total; i++ {
		shortID := fmt.Sprintf("lim%02d", i)
		entries = append(entries, hottestEntry{ShortID: shortID, CreatedAt: createdAt})
	}
	rt := newStubRoundTripper()
	rt.setHottest(entries...)
	for _, e := range entries {
		rt.setStory(e.ShortID, storyJSON(e.ShortID, createdAt, "u", "story", "本文", "", "https://lobste.rs/s/"+e.ShortID, ""))
	}
	source := newStubListItemSource(rt)

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と fetch 回数
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(got) != lobsters.MaxStoriesScanned {
		t.Fatalf("len(got) = %d, want %d", len(got), lobsters.MaxStoriesScanned)
	}
	// 上限件数に達した後、次の story は fetch しない。
	droppedID := entries[lobsters.MaxStoriesScanned].ShortID
	if rt.storyCalls(droppedID) != 0 {
		t.Fatalf("story %q fetched after limit reached (calls=%d)", droppedID, rt.storyCalls(droppedID))
	}
}

func TestList_retriesOnceOnTransientError_whenSecondAttemptSucceeds(t *testing.T) {
	// @given hottest.json が 1 回目 5xx、2 回目成功を返す double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(time.Hour).Format(time.RFC3339)
	hottestCalls := 0
	transientRT := &sequenceRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/hottest.json" {
				hottestCalls++
				if hottestCalls == 1 {
					return &http.Response{
						StatusCode: http.StatusBadGateway,
						Body:       io.NopCloser(strings.NewReader("boom")),
						Header:     make(http.Header),
						Request:    req,
					}, nil
				}
				body := fmt.Sprintf(`[{"short_id":"retry1","created_at":%q}]`, createdAt)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}
			if strings.HasPrefix(req.URL.Path, "/s/") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(storyJSON(
						"retry1", createdAt, "u", "再試行成功", "本文", "",
						"https://lobste.rs/s/retry1", "",
					))),
					Header:  make(http.Header),
					Request: req,
				}, nil
			}
			return nil, fmt.Errorf("sequenceRoundTripper: unexpected path %q", req.URL.Path)
		},
	}
	source := lobsters.NewListItemSource(&http.Client{Transport: transientRT})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と hottest fetch 回数
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if hottestCalls != 2 {
		t.Fatalf("hottest fetch attempts = %d, want 2 (retry once)", hottestCalls)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (%+v)", len(got), got)
	}
}

func TestList_parsesCreatedAtWithTimezoneOffset_toUTC(t *testing.T) {
	// @given created_at が +09:00 offset 付きの double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := "2026-09-01T12:00:00+09:00"
	wantOccurredAt := time.Date(2026, 9, 1, 3, 0, 0, 0, time.UTC)
	rt := newStubRoundTripper()
	rt.setHottest(hottestEntry{ShortID: "tz1", CreatedAt: createdAt})
	rt.setStory("tz1", storyJSON("tz1", createdAt, "u", "TZ テスト", "", "", "https://lobste.rs/s/tz1", ""))
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

// sequenceRoundTripper は request ごとに任意の応答を返す最小 Transport。
type sequenceRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (rt *sequenceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.handler(req)
}
