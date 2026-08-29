package getxapi_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

const getXAPITestAPIKey = "getxapi-test-real-value"

// newTestHTTPClient は固定 upstream URL への接続を test TLS server へ差し替えた *http.Client を返す。
// why: Adapter は本番 host を定数として持つため、DialTLSContext で接続先だけを test server へ redirect する。
func newTestHTTPClient(server *httptest.Server) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
}

func TestList_returnsSourceItems_forAllWatchedUsers(t *testing.T) {
	// Given: watch user と vendor response を返す upstream double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	var gotAuthorization string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{"tweets":[{"id":"tweet-1","url":"https://x.example/tweet-1","text":"本文","createdAt":"Wed Aug 19 10:00:00 +0000 2026","author":{"id":"author-1","name":"表示名"},"entities":{"urls":[{"expanded_url":"https://example.com"}]},"media":[{"url":"https://img.example/a.jpg"}]}],"has_more":false}`)
	}))
	t.Cleanup(server.Close)
	source := getxapi.NewPostSource(newTestHTTPClient(server), getXAPITestAPIKey)
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: SourceItem と Context の規約を満たし、Bearer に API key が乗る
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotAuthorization != "Bearer "+getXAPITestAPIKey {
		t.Fatalf("Authorization = %q, want %q", gotAuthorization, "Bearer "+getXAPITestAPIKey)
	}
	wantOccurredAt := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	wantContext := strings.Join([]string{
		"item_id: tweet-1",
		"actor_id: author-1",
		"actor_name: 表示名",
		"text: 本文",
		"permalink: https://x.example/tweet-1",
		"links: https://example.com",
		"media: https://img.example/a.jpg",
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
}

func TestList_paginates_and_stops_atSince(t *testing.T) {
	// Given: 1ページ目は次ページを持ち、2ページ目は since より古い
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			_, _ = io.WriteString(w, `{"tweets":[{"id":"new","createdAt":"Wed Aug 19 10:00:00 +0000 2026"}],"has_more":true,"next_cursor":"next"}`)
			return
		}
		_, _ = io.WriteString(w, `{"tweets":[{"id":"old","createdAt":"Tue Aug 18 10:00:00 +0000 2026"}],"has_more":false}`)
	}))
	t.Cleanup(server.Close)
	source := getxapi.NewPostSource(newTestHTTPClient(server), getXAPITestAPIKey)

	// When: since を指定して List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC))

	// Then: 2ページ目まで取得し、古い要素は含めない
	if err != nil || len(got) != 1 || calls != 2 {
		t.Fatalf("got = %+v, err = %v, calls = %d", got, err, calls)
	}
}

func TestList_returnsInfrastructureError_whenClientNil(t *testing.T) {
	// Given: client が nil
	source := getxapi.NewPostSource(nil, getXAPITestAPIKey)
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: Infrastructure Error。Error / Unwrap が観測できる
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *getxapi.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *getxapi.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "getxapi:") {
		t.Fatalf("Error() = %q, want prefix getxapi:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestList_returnsError_whenResponseBodyIsInvalidJSON(t *testing.T) {
	// Given: watch user と、200 だが JSON でない body を返す upstream double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "not-json")
	}))
	t.Cleanup(server.Close)
	source := getxapi.NewPostSource(newTestHTTPClient(server), getXAPITestAPIKey)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: decode 失敗で error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
	var infra *getxapi.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *getxapi.Error", err, err)
	}
}

func TestList_returnsError_whenConnectionFailsMidRequest(t *testing.T) {
	// Given: watch user と、接続直後に閉じる upstream
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := server.Listener.Addr().String()
	server.Close()
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addrIgnored string) (net.Conn, error) {
				return tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	source := getxapi.NewPostSource(httpClient, getXAPITestAPIKey)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: do 失敗で error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
}

func TestList_returnsError_whenVendorStatusIsNotOK(t *testing.T) {
	// Given: watch user と HTTP error を返す upstream double
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "失敗", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	source := getxapi.NewPostSource(newTestHTTPClient(server), getXAPITestAPIKey)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: 部分結果なしで error
	if got != nil || err == nil {
		t.Fatalf("got = %+v, err = %v, want nil and error", got, err)
	}
}
