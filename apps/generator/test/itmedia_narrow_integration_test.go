// Scope: Narrow Integration
// 実物境界: itmedia.ListItemSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 ITmedia RSS は使わない。DialTLSContext で本番 host（rss.itmedia.co.jp）宛先だけを test server へ redirect する。
// @require upstream は controllable な test server。ITmedia RSS は認証不要のため secret を渡さない。
// @ensure upstream は GET を受け取り、Authorization header は空（認証 header 無しで成功する）。
// @ensure 成功時 List は SourceItem を 1 件以上返す。
// @ensure 失敗経路（5xx を返す upstream）で *itmedia.Error が返る。
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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/itmedia"
)

// itmediaNarrowProbe は upstream が受けた request の観測面を記録する。
type itmediaNarrowProbe struct {
	methods        []string
	authorizations []string
	paths          []string
}

// newITmediaListItemSourceWithProxy は本番 host への接続を test TLS server へ redirect した ListItemSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure 標準 *http.Client がそのまま（認証 header なしで）request を送る。
func newITmediaListItemSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*itmedia.ListItemSource, *itmediaNarrowProbe) {
	t.Helper()
	probe := &itmediaNarrowProbe{}
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
				if host != "rss.itmedia.co.jp" {
					return nil, fmt.Errorf("unexpected TLS host %q", host)
				}
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
	return itmedia.NewListItemSource(httpClient), probe
}

func itmediaRSSFixture(pubDate, title, link, description string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<rss version="2.0">` + "\n" +
		"<channel>\n" +
		fmt.Sprintf("<item><title>%s</title><link>%s</link><description>%s</description><pubDate>%s</pubDate></item>\n", title, link, description, pubDate) +
		"</channel>\n</rss>\n"
}

func TestITmediaListItemSource_deliversGetWithoutAuthHeader_whenUpstreamSucceeds(t *testing.T) {
	// @given 成功応答を返す upstream double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	occurredAt := since.Add(2 * time.Hour)
	pubDate := occurredAt.Format(time.RFC1123Z)
	source, probe := newITmediaListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/rss/2.0/news_bursts.xml") {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, itmediaRSSFixture(
			pubDate,
			"Narrow 記事",
			"https://www.itmedia.co.jp/news/articles/narrow.html",
			"本文",
		))
	})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と upstream 観測面
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
	if len(probe.paths) == 0 {
		t.Fatal("probe.paths is empty, want feed path")
	}
	if !strings.HasSuffix(probe.paths[0], "/rss/2.0/news_bursts.xml") {
		t.Fatalf("paths[0] = %q, want suffix %q", probe.paths[0], "/rss/2.0/news_bursts.xml")
	}
	if len(got) < 1 {
		t.Fatalf("len(got) = %d, want >= 1", len(got))
	}
	if got[0].SourceID != itmedia.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, itmedia.SourceID)
	}
}

func TestITmediaListItemSource_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	// @given feed が常に 502 を返す upstream double
	since := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	source, probe := newITmediaListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	// @when
	got, err := source.List(context.Background(), since)

	// @then 戻り値と再試行回数
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *itmedia.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *itmedia.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "itmedia:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "itmedia:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
	if len(probe.methods) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(probe.methods))
	}
}
