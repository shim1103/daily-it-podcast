// Scope: Narrow Integration
// 実物境界: publickey.ListItemSource が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 Publickey 実 feed は使わない。DialTLSContext で本番 host（www.publickey1.jp）宛先だけを test server へ redirect する。
// @require upstream は controllable な test server。Atom feed は認証不要のため secret を渡さない。
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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/publickey"
)

// publickeyNarrowProbe は upstream が受けた request の観測面を記録する。
type publickeyNarrowProbe struct {
	methods        []string
	authorizations []string
	paths          []string
}

// newPublickeyListItemSourceWithProxy は本番 host への接続を test TLS server へ redirect した ListItemSource を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure 標準 *http.Client がそのまま（認証 header なしで）request を送る。
func newPublickeyListItemSourceWithProxy(t *testing.T, handler http.HandlerFunc) (*publickey.ListItemSource, *publickeyNarrowProbe) {
	t.Helper()
	probe := &publickeyNarrowProbe{}
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
				if host != "www.publickey1.jp" {
					return nil, fmt.Errorf("unexpected TLS host %q", host)
				}
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
	return publickey.NewListItemSource(httpClient, publickey.MaxStoriesScanned, &retryReporterSpy{}), probe
}

func TestPublickeyListItemSource_deliversGetWithoutAuthHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: /atom.xml が window 内 entry 1 件を返す httptest TLS upstream（本番 Publickey は使わない）
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	published := since.Add(2 * time.Hour).UTC().Format(time.RFC3339)
	source, probe := newPublickeyListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/atom.xml" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
				`<feed xmlns="http://www.w3.org/2005/Atom">`+"\n"+
				`<title>Publickey</title>`+"\n"+
				`<entry>`+"\n"+
				`<title>Narrow 記事</title>`+"\n"+
				`<link rel="alternate" type="text/html" href="https://www.publickey1.jp/blog/26/narrow.html" />`+"\n"+
				`<id>tag:www.publickey1.jp,2026://2.narrow</id>`+"\n"+
				`<published>%s</published>`+"\n"+
				`<summary>要約</summary>`+"\n"+
				`<content type="html">本文</content>`+"\n"+
				`<author><name>narrowuser</name></author>`+"\n"+
				`</entry></feed>`,
			published,
		))
	})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: upstream は GET・Authorization 空、戻りは SourceID=publickey の 1 件以上
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
	if got[0].SourceID != publickey.SourceID {
		t.Fatalf("SourceID = %q, want %q", got[0].SourceID, publickey.SourceID)
	}
}

func TestPublickeyListItemSource_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	// Given: /atom.xml が常に 502 を返す httptest TLS upstream
	since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	source, probe := newPublickeyListItemSourceWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: *adaptererror.Error（publickey: prefix）かつ feed 5xx で 2 回 request（retry once）
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "publickey:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "publickey:")
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() = nil, want non-nil")
	}
	if len(probe.methods) != 2 {
		t.Fatalf("upstream received %d requests, want 2 (retry once on 5xx)", len(probe.methods))
	}
}
