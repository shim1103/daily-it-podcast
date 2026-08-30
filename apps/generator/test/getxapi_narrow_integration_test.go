// Scope: Narrow Integration
// 実物境界: getxapi.PostSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 credential / GetXAPI 実 API は使わない。DialTLSContext で本番 host 宛先だけを test server へ redirect する。
// @require dummy API key を Adapter へ直接渡す。upstream は controllable な test server。
// @ensure upstream は GET を受け取り、Authorization header に Bearer が届く。
// @ensure 成功時 List は SourceItem を 1 件以上返す。
// @invariant dummy secret 実値は error message へ出ない。
package test

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

// newGetXAPIPostSourceWithProxy は本番 endpoint への接続を test server へ redirect した PostSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure dummy API key は Adapter へ直接渡し、標準 *http.Client がそのまま Authorization Bearer へ乗せる。
func newGetXAPIPostSourceWithProxy(t *testing.T, apiKey string, handler http.HandlerFunc) *getxapi.PostSource {
	t.Helper()

	upstream := httptest.NewTLSServer(handler)
	t.Cleanup(upstream.Close)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	return getxapi.NewPostSource(httpClient, apiKey)
}

func TestGetXAPIPostSource_deliversGetWithBearerAuthorization_whenUpstreamSucceeds(t *testing.T) {
	// Given: watch user、dummy API key、成功応答を返す upstream double
	const apiKey = "narrow-getxapi-real-value"
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })

	var gotMethod string
	var gotAuthorization string
	source := newGetXAPIPostSourceWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuthorization = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{"tweets":[{"id":"tweet-1","url":"https://x.example/tweet-1","text":"本文","createdAt":"Wed Aug 19 10:00:00 +0000 2026","author":{"id":"author-1","name":"表示名"},"entities":{"urls":[{"expanded_url":"https://example.com"}]},"media":[{"url":"https://img.example/a.jpg"}]}],"has_more":false}`)
	})
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List する
	got, err := source.List(context.Background(), since)

	// Then: upstream は GET を受け、Bearer が届き、SourceItem が 1 件以上返る
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("method = %q, want %q", gotMethod, http.MethodGet)
	}
	if gotAuthorization != "Bearer "+apiKey {
		t.Fatalf("Authorization = %q, want Bearer prefix with configured key", gotAuthorization)
	}
	if len(got) < 1 {
		t.Fatalf("len(got) = %d, want >= 1", len(got))
	}
}

func TestGetXAPIPostSource_excludesDummySecretFromErrorMessage_whenUpstreamFails(t *testing.T) {
	// Given: watch user、dummy API key、常に 502 を返す upstream double
	const apiKey = "narrow-getxapi-must-not-leak-value"
	previous := x.WatchUserIDs
	x.WatchUserIDs = []string{"user-1"}
	t.Cleanup(func() { x.WatchUserIDs = previous })

	var gotMethod string
	source := newGetXAPIPostSourceWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		http.Error(w, "失敗", http.StatusBadGateway)
	})

	// When: List する
	got, err := source.List(context.Background(), time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC))

	// Then: error は返るが、dummy secret 実値は error message に出ない
	if gotMethod == "" {
		t.Fatal("upstream was not called")
	}
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	if err == nil {
		t.Fatal("List() error = nil, want non-nil")
	}
	var infra *getxapi.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *getxapi.Error", err, err)
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("error message contains dummy secret value")
	}
}
