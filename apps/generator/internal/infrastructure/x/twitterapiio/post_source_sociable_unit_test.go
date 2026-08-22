package twitterapiio_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/twitterapiio"
)

func TestList_returnsSourceItems_withoutMediaContext(t *testing.T) {
	// Given: watch user と vendor response を返す proxy double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"tweets":[{"id":"tweet-1","url":"https://x.example/tweet-1","text":"本文","createdAt":"Wed Aug 19 10:00:00 +0000 2026","author":{"id":"author-1","userName":"user_name"},"entities":{"urls":[{"expanded_url":"https://example.com"}]}}],"has_next_page":false,"status":"success"}`)
	}))
	t.Cleanup(server.Close)
	source := twitterapiio.NewPostSource(&agentsecrets.Client{HTTP: server.Client(), ProxyURL: server.URL})
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: media 行なしの SourceItem
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	wantOccurredAt := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	wantContext := strings.Join([]string{
		"item_id: tweet-1",
		"actor_id: author-1",
		"actor_name: user_name",
		"text: 本文",
		"permalink: https://x.example/tweet-1",
		"links: https://example.com",
	}, "\n")
	if len(got) != 1 {
		t.Fatalf("got = %+v, want one SourceItem", got)
	}
	if got[0].SourceID != x.SourceID ||
		!got[0].OccurredAt.Equal(wantOccurredAt) ||
		got[0].OccurredAt.Location() != time.UTC ||
		got[0].Context != wantContext {
		t.Fatalf("got[0] = %+v, want SourceID=%q OccurredAt=%v Context=%q", got[0], x.SourceID, wantOccurredAt, wantContext)
	}
	if strings.Contains(got[0].Context, "media:") {
		t.Fatalf("Context = %q, must not contain media", got[0].Context)
	}
}

func TestList_paginates_and_stops_atSince(t *testing.T) {
	// Given: 1ページ目は次ページを持ち、2ページ目は since より古い
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_, _ = io.WriteString(w, `{"tweets":[{"id":"new","createdAt":"Wed Aug 19 10:00:00 +0000 2026"}],"has_next_page":true,"next_cursor":"next","status":"success"}`)
			return
		}
		_, _ = io.WriteString(w, `{"tweets":[{"id":"old","createdAt":"Tue Aug 18 10:00:00 +0000 2026"}],"has_next_page":false,"status":"success"}`)
	}))
	t.Cleanup(server.Close)
	source := twitterapiio.NewPostSource(&agentsecrets.Client{HTTP: server.Client(), ProxyURL: server.URL})

	// When: since を指定して List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC))

	// Then: 2ページ目まで取得し、古い要素は含めない
	if err != nil || len(got) != 1 || calls != 2 {
		t.Fatalf("got = %+v, err = %v, calls = %d", got, err, calls)
	}
}

func TestList_returnsInfrastructureError_whenClientNil(t *testing.T) {
	// Given: client が nil
	source := twitterapiio.NewPostSource(nil)
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: Infrastructure Error。Error / Unwrap が観測できる
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
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

func TestList_returnsError_whenVendorStatusIsError(t *testing.T) {
	// Given: watch user と vendor error response を返す proxy double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"error","message":"失敗"}`)
	}))
	t.Cleanup(server.Close)
	source := twitterapiio.NewPostSource(&agentsecrets.Client{HTTP: server.Client(), ProxyURL: server.URL})

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: 部分結果なしで error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
}
