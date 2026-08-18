package twitterapiio_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/twitterapiio"
)

const (
	watchUserID     = "44196397"
	createdAtLayout = "Mon Jan 02 15:04:05 -0700 2006"
)

type proxyProbe struct {
	TargetURLs []string
	Bearers    []string
	APIKeys    []string
}

func newPostSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*twitterapiio.PostSource, *proxyProbe) {
	t.Helper()
	probe := &proxyProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probe.TargetURLs = append(probe.TargetURLs, r.Header.Get("X-AS-Target-URL"))
		probe.Bearers = append(probe.Bearers, r.Header.Get("X-AS-Inject-Bearer"))
		probe.APIKeys = append(probe.APIKeys, r.Header.Get("X-AS-Inject-Header-X-API-Key"))
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	source := twitterapiio.NewPostSource(&agentsecrets.Client{
		HTTP:     server.Client(),
		ProxyURL: server.URL,
	})
	return source, probe
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

func lastTweetsPage(tweets []map[string]any, nextCursor string, hasNext bool) map[string]any {
	return map[string]any{
		"tweets":        tweets,
		"has_next_page": hasNext,
		"next_cursor":   nextCursor,
		"status":        "success",
		"message":       "success",
	}
}

func originalTweet(id, createdAt string) map[string]any {
	return map[string]any{
		"id":        id,
		"url":       "https://x.com/elonmusk/status/" + id,
		"text":      "本文-" + id,
		"createdAt": createdAt,
		"isReply":   false,
		"author":    map[string]any{"id": watchUserID, "userName": "elonmusk"},
		"entities": map[string]any{
			"urls": []map[string]any{
				{"expanded_url": "https://example.com/" + id, "url": "https://t.co/" + id},
			},
		},
	}
}

func queryOf(t *testing.T, targetURL string) url.Values {
	t.Helper()
	u, err := url.Parse(targetURL)
	if err != nil {
		t.Fatalf("parse target URL: %v", err)
	}
	return u.Query()
}

func TestListByUser_returnsOnlyOriginalPosts_whenReplyRepostAndQuotePresent(t *testing.T) {
	t.Parallel()

	// Given: vendor 1 page に original / Reply / Repost / 引用が混在する
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	createdAt := since.Format(createdAtLayout)
	source, _ := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		reply := originalTweet("reply-1", createdAt)
		reply["isReply"] = true
		repost := originalTweet("repost-1", createdAt)
		repost["retweeted_tweet"] = map[string]any{"id": "orig-rt"}
		quote := originalTweet("quote-1", createdAt)
		quote["quoted_tweet"] = map[string]any{"id": "orig-qt"}
		writeJSON(t, w, lastTweetsPage([]map[string]any{
			originalTweet("orig-1", createdAt),
			reply,
			repost,
			quote,
		}, "", false))
	})

	// When: ListByUser を呼ぶ
	got, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: original のみ返す
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].ID != "orig-1" {
		t.Fatalf("ID = %q, want orig-1", got[0].ID)
	}
}

func TestListByUser_includesSinceBoundaryAndExcludesOlder_whenCreatedAtAroundSince(t *testing.T) {
	t.Parallel()

	// Given: since ちょうどと since より前の original が 1 page に並ぶ
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	onSince := "Tue Dec 10 00:00:00 +0000 2024"
	beforeSince := "Mon Dec 09 23:59:59 +0000 2024"
	source, _ := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, lastTweetsPage([]map[string]any{
			originalTweet("on-since", onSince),
			originalTweet("before-since", beforeSince),
		}, "", false))
	})

	// When: ListByUser を呼ぶ
	got, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: CreatedAt >= since のみ。境界値は含む
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].ID != "on-since" {
		t.Fatalf("ID = %q, want on-since", got[0].ID)
	}
	if got[0].CreatedAt.Before(since) {
		t.Fatalf("CreatedAt %v is before since %v", got[0].CreatedAt, since)
	}
}

func TestListByUser_returnsEmptySlice_whenNoOriginalInWindow(t *testing.T) {
	t.Parallel()

	// Given: window 内 original が 0 件の成功応答
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	source, _ := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, lastTweetsPage(nil, "", false))
	})

	// When: ListByUser を呼ぶ
	got, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: 空 slice（nil ではない）
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if got == nil {
		t.Fatal("got nil slice, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestListByUser_mapsDocumentedFieldsToPost_whenOriginalTweet(t *testing.T) {
	t.Parallel()

	// Given: OpenAPI Tweet の documented field を持つ original 1 件
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	createdAt := "Tue Dec 10 07:00:30 +0000 2024"
	source, _ := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, lastTweetsPage([]map[string]any{
			originalTweet("1850000000000000000", createdAt),
		}, "", false))
	})

	// When: ListByUser を呼ぶ
	got, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: models.Post の field だけが埋まる
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	wantTime, err := time.Parse(createdAtLayout, createdAt)
	if err != nil {
		t.Fatalf("parse want time: %v", err)
	}
	post := got[0]
	if post.ID != "1850000000000000000" {
		t.Fatalf("ID = %q", post.ID)
	}
	if post.AuthorID != watchUserID {
		t.Fatalf("AuthorID = %q", post.AuthorID)
	}
	if post.Text != "本文-1850000000000000000" {
		t.Fatalf("Text = %q", post.Text)
	}
	if !post.CreatedAt.Equal(wantTime) {
		t.Fatalf("CreatedAt = %v, want %v", post.CreatedAt, wantTime)
	}
	if post.Permalink != "https://x.com/elonmusk/status/1850000000000000000" {
		t.Fatalf("Permalink = %q", post.Permalink)
	}
	if len(post.URLs) != 1 || post.URLs[0] != "https://example.com/1850000000000000000" {
		t.Fatalf("URLs = %#v", post.URLs)
	}
	if post.Media == nil {
		t.Fatal("Media is nil, want empty slice")
	}
	if len(post.Media) != 0 {
		t.Fatalf("Media = %#v, want empty", post.Media)
	}
}

func TestListByUser_injectsTwitterIOAPIKeyName_whenCallingProxy(t *testing.T) {
	t.Parallel()

	// Given: 成功する空 page を返す proxy double
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	source, probe := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, lastTweetsPage(nil, "", false))
	})

	// When: ListByUser を呼ぶ
	_, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: X-AS-Inject-Header-X-API-Key にキー名だけが渡り、Bearer は付かない
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(probe.APIKeys) != 1 {
		t.Fatalf("API key headers = %#v", probe.APIKeys)
	}
	if probe.APIKeys[0] != "TWITTER_IO_API_KEY" {
		t.Fatalf("X-AS-Inject-Header-X-API-Key = %q, want TWITTER_IO_API_KEY", probe.APIKeys[0])
	}
	if len(probe.Bearers) != 1 {
		t.Fatalf("bearer headers = %#v", probe.Bearers)
	}
	if probe.Bearers[0] != "" {
		t.Fatalf("X-AS-Inject-Bearer = %q, want empty", probe.Bearers[0])
	}
	if len(probe.TargetURLs) != 1 {
		t.Fatalf("target URLs = %#v", probe.TargetURLs)
	}
	q := queryOf(t, probe.TargetURLs[0])
	if q.Get("userId") != watchUserID {
		t.Fatalf("userId = %q", q.Get("userId"))
	}
	if q.Get("includeReplies") != "false" {
		t.Fatalf("includeReplies = %q, want false", q.Get("includeReplies"))
	}
	target := probe.TargetURLs[0]
	if target == "" {
		t.Fatal("X-AS-Target-URL is empty")
	}
}

func TestListByUser_followsCursorThenStops_whenOlderThanSinceAppears(t *testing.T) {
	t.Parallel()

	// Given: 1 page 目は window 内、2 page 目に since より前が現れ、3 page 目もある
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	source, probe := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-AS-Target-URL")
		u, err := url.Parse(target)
		if err != nil {
			t.Fatalf("parse target: %v", err)
		}
		switch u.Query().Get("cursor") {
		case "":
			writeJSON(t, w, lastTweetsPage([]map[string]any{
				originalTweet("new-1", "Wed Dec 11 00:00:00 +0000 2024"),
			}, "cursor-2", true))
		case "cursor-2":
			writeJSON(t, w, lastTweetsPage([]map[string]any{
				originalTweet("old-1", "Mon Dec 09 00:00:00 +0000 2024"),
			}, "cursor-3", true))
		default:
			t.Fatalf("unexpected cursor %q", u.Query().Get("cursor"))
		}
	})

	// When: ListByUser を呼ぶ
	got, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: window 内だけ返し、since より前の page の次は呼ばない
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(got) != 1 || got[0].ID != "new-1" {
		t.Fatalf("got = %+v, want [new-1]", got)
	}
	if len(probe.TargetURLs) != 2 {
		t.Fatalf("request count = %d, want 2: %#v", len(probe.TargetURLs), probe.TargetURLs)
	}
	if queryOf(t, probe.TargetURLs[1]).Get("cursor") != "cursor-2" {
		t.Fatalf("second cursor = %q", queryOf(t, probe.TargetURLs[1]).Get("cursor"))
	}
}

func TestListByUser_returnsInfrastructureError_whenProxyStatusNotOK(t *testing.T) {
	t.Parallel()

	// Given: proxy が 400 を返す
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	source, _ := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":400,"message":"bad"}`)
	})

	// When: ListByUser を呼ぶ
	_, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: Infrastructure Error が返る
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *twitterapiio.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *twitterapiio.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "twitterapiio:") {
		t.Fatalf("Error() = %q, want prefix twitterapiio:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestListByUser_returnsInfrastructureError_whenVendorStatusIsError(t *testing.T) {
	t.Parallel()

	// Given: HTTP 200 だが status=error の body
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)
	source, _ := newPostSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"tweets":        []any{},
			"has_next_page": false,
			"next_cursor":   "",
			"status":        "error",
			"message":       "upstream failed",
		})
	})

	// When: ListByUser を呼ぶ
	_, err := source.ListByUser(context.Background(), watchUserID, since)

	// Then: Infrastructure Error が返る
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *twitterapiio.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *twitterapiio.Error", err, err)
	}
}
