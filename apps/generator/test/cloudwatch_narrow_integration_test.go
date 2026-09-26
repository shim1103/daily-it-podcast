// Scope: Narrow Integration
// 実物境界: cloudwatch.ListItemSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番クラウド Watch 実 feed は使わない。DialTLSContext で本番 host（cloud.watch.impress.co.jp）宛先だけを test server へ redirect する。
// @require upstream は controllable な test server。RDF feed は認証不要のため secret を渡さない。
// @ensure upstream は GET を受け取り、Authorization header は空（認証 header 無しで成功する）。
// @ensure 成功時 List は SourceItem を 1 件以上返す。
// @ensure 失敗経路（5xx を返す upstream）で *adaptererror.Error が返る。
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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/cloudwatch"
)

// cloudwatchNarrowProbe は upstream が受けた request の観測面を記録する。
type cloudwatchNarrowProbe struct {
	methods        []string
	authorizations []string
	paths          []string
}

// newCloudWatchListItemSourceWithProxy は本番 host への接続を test TLS server へ redirect した ListItemSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure 標準 *http.Client がそのまま（認証 header なしで）request を送る。
func newCloudWatchListItemSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*cloudwatch.ListItemSource, *cloudwatchNarrowProbe) {
	t.Helper()
	probe := &cloudwatchNarrowProbe{}
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
				if host != "cloud.watch.impress.co.jp" {
					return nil, fmt.Errorf("unexpected TLS host %q", host)
				}
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
	return cloudwatch.NewListItemSource(httpClient, cloudwatch.MaxStoriesScanned), probe
}

func TestCloudWatchListItemSource_deliversGetWithoutAuthHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: /data/rss/1.0/clw/feed.rdf が window 内 RDF item 1 件を返す httptest TLS upstream（本番クラウド Watch は使わない）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	date := since.Add(2 * time.Hour).UTC().Format(time.RFC3339)
	source, probe := newCloudWatchListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/rss/1.0/clw/feed.rdf" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
				`<rdf:RDF xmlns="http://purl.org/rss/1.0/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:dc="http://purl.org/dc/elements/1.1/" xml:lang="ja">`+"\n"+
				`<channel rdf:about="https://cloud.watch.impress.co.jp/data/rss/1.0/clw/feed.rdf">`+"\n"+
				`<title>クラウド Watch</title>`+"\n"+
				`</channel>`+"\n"+
				`<item rdf:about="https://cloud.watch.impress.co.jp/docs/news/narrow.html?ref=rss">`+"\n"+
				`<title>Narrow 記事</title>`+"\n"+
				`<link>https://cloud.watch.impress.co.jp/docs/news/narrow.html</link>`+"\n"+
				`<dc:date>%s</dc:date>`+"\n"+
				`<description><![CDATA[説明]]></description>`+"\n"+
				`</item></rdf:RDF>`,
			date,
		))
	})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: upstream は GET・Authorization 空、戻りは SourceID=cloudwatch の 1 件以上
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
	if got[0].SourceID != cloudwatch.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, cloudwatch.SourceID)
	}
}

func TestCloudWatchListItemSource_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	// Given: feed.rdf が常に 502 を返す httptest TLS upstream
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	source, probe := newCloudWatchListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: *adaptererror.Error（cloudwatch: prefix）かつ feed 5xx で 2 回 request（retry once）
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "cloudwatch:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "cloudwatch:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
	if len(probe.methods) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(probe.methods))
	}
}
