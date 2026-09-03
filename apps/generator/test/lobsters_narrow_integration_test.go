// Scope: Narrow Integration
// 実物境界: lobsters.ListItemSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 Lobsters 実 API は使わない。DialTLSContext で本番 host（lobste.rs）宛先だけを test server へ redirect する。
// @require upstream は controllable な test server。Lobsters JSON は認証不要のため secret を渡さない。
// @ensure upstream は GET を受け取り、Authorization header は空（認証 header 無しで成功する）。
// @ensure 成功時 List は SourceItem を 1 件以上返す。
// @ensure 失敗経路（5xx を返す upstream）で *lobsters.Error が返る。
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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/lobsters"
)

// lobstersNarrowProbe は upstream が受けた request の観測面を記録する。
type lobstersNarrowProbe struct {
	methods        []string
	authorizations []string
	paths          []string
}

// newLobstersListItemSourceWithProxy は本番 host への接続を test TLS server へ redirect した ListItemSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure 標準 *http.Client がそのまま（認証 header なしで）request を送る。
func newLobstersListItemSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*lobsters.ListItemSource, *lobstersNarrowProbe) {
	t.Helper()
	probe := &lobstersNarrowProbe{}
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
				if host != "lobste.rs" {
					return nil, fmt.Errorf("unexpected TLS host %q", host)
				}
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
	return lobsters.NewListItemSource(httpClient), probe
}

func TestLobstersListItemSource_deliversGetWithoutAuthHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: hottest→story 詳細の成功応答を返す upstream double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	createdAt := since.Add(2 * time.Hour).Format(time.RFC3339)
	source, probe := newLobstersListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/hottest.json"):
			_, _ = io.WriteString(w, fmt.Sprintf(`[{"short_id":"narrow1","created_at":%q}]`, createdAt))
		case strings.HasSuffix(r.URL.Path, "/s/narrow1.json"):
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"short_id":"narrow1","submitter_user":"narrowuser","title":"Narrow 記事","description_plain":"本文","url":"https://example.com/n","short_id_url":"https://lobste.rs/s/narrow1","comments_url":"https://lobste.rs/s/narrow1/comments","created_at":%q,"comments":[]}`,
				createdAt,
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
	if got[0].SourceID != lobsters.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, lobsters.SourceID)
	}
}

func TestLobstersListItemSource_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	// Given: hottest.json が常に 502 を返す upstream double
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	source, probe := newLobstersListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	// When: List する
	got, err := source.List(context.Background(), since)

	// Then: 戻り値と再試行回数
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
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
	if len(probe.methods) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(probe.methods))
	}
}
