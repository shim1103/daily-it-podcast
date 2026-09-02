// Scope: Narrow Integration
// 実物境界: hackernews.ListItemSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 HackerNews 実 API は使わない。DialTLSContext で本番 host（hacker-news.firebaseio.com）宛先だけを test server へ redirect する。
// @require upstream は controllable な test server。HackerNews は認証不要のため secret を渡さない。
// @ensure upstream は GET を受け取り、Authorization header は空（認証 header 無しで成功する）。
// @ensure 成功時 List は SourceItem を 1 件以上返す。
// @ensure 失敗経路（5xx を返す upstream）で *hackernews.Error が返る。
// @invariant vendor 固有型・監視対象一覧を露出しない。
package test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
)

// hackernewsNarrowProbe は upstream が受けた request の観測面を記録する。
type hackernewsNarrowProbe struct {
	methods        []string
	authorizations []string
	paths          []string
}

// newHackerNewsListItemSourceWithProxy は本番 host への接続を test TLS server へ redirect した ListItemSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure 標準 *http.Client がそのまま（認証 header なしで）request を送る。
func newHackerNewsListItemSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*hackernews.ListItemSource, *hackernewsNarrowProbe) {
	t.Helper()
	probe := &hackernewsNarrowProbe{}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probe.methods = append(probe.methods, r.Method)
		probe.authorizations = append(probe.authorizations, r.Header.Get("Authorization"))
		probe.paths = append(probe.paths, r.URL.Path)
		handler(w, r)
	}))
	t.Cleanup(upstream.Close)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				if host != "hacker-news.firebaseio.com" {
					return nil, fmt.Errorf("unexpected TLS host %q", host)
				}
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
	return hackernews.NewListItemSource(httpClient), probe
}

func TestHackerNewsListItemSource_deliversGetWithoutAuthHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: topstories→item の成功応答を返す upstream double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	storyUnix := since.Add(2 * time.Hour).Unix()
	source, probe := newHackerNewsListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/topstories.json"):
			_, _ = io.WriteString(w, "[501]")
		case strings.HasSuffix(r.URL.Path, "/item/501.json"):
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"id":501,"type":"story","by":"narrowuser","time":%d,"title":"Narrow 記事","text":"本文","url":"https://example.com/n","kids":[]}`,
				storyUnix,
			))
		default:
			http.Error(w, "unexpected path", http.StatusNotFound)
		}
	})

	// When: List する
	got, err := source.List(context.Background(), since)

	// Then: 戻り値と upstream 観測面
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(probe.methods) == 0 {
		t.Fatal("upstream was not called")
	}
	for i, m := range probe.methods {
		if m != http.MethodGet {
			t.Fatalf("methods[%d] = %q, want %q", i, m, http.MethodGet)
		}
	}
	for i, auth := range probe.authorizations {
		if auth != "" {
			t.Fatalf("authorizations[%d] = %q, want empty (no auth header)", i, auth)
		}
	}
	if len(got) < 1 {
		t.Fatalf("len(got) = %d, want >= 1", len(got))
	}
	if got[0].SourceID != hackernews.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, hackernews.SourceID)
	}
}

func TestHackerNewsListItemSource_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	// Given: topstories.json が常に 502 を返す upstream double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	source, probe := newHackerNewsListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	// When: List する
	got, err := source.List(context.Background(), since)

	// Then: 戻り値と再試行回数
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *hackernews.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *hackernews.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "hackernews:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "hackernews:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
	if len(probe.methods) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(probe.methods))
	}
}
