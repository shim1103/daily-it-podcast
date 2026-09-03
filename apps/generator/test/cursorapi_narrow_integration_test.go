// Scope: Narrow Integration
// 実物境界: cursorapi.TextWriter が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 credential / Cursor 実 API は使わない。DialTLSContext で本番 host（api.cursor.com）宛先だけを test server へ redirect する。
// @require dummy API key を Adapter へ直接渡す。upstream は controllable な test server。
// @ensure upstream は create の POST と stream の GET を受け取り、Authorization: Bearer に実値が届く。
// @ensure 成功時 Write は非空断片を返す。
// @invariant dummy secret 実値は error message へ出ない。
package test

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorapi"
)

type cursorNarrowProbe struct {
	createMethod string
	createAuth   string
	streamMethod string
	streamPath   string
}

// newCursorTextWriterWithProxy は本番 host（api.cursor.com）への接続を test TLS server へ差し替えた TextWriter を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure dummy API key は Adapter へ直接渡し、標準 *http.Client が Authorization header へ載せる。
func newCursorTextWriterWithProxy(t *testing.T, apiKey string, handler http.HandlerFunc) (*cursorapi.TextWriter, *cursorNarrowProbe) {
	t.Helper()
	probe := &cursorNarrowProbe{}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/v1/agents"):
			probe.createMethod = r.Method
			probe.createAuth = r.Header.Get("Authorization")
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/runs/"):
			probe.streamMethod = r.Method
			probe.streamPath = r.URL.Path
		}
		handler(w, r)
	}))
	t.Cleanup(upstream.Close)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true}) //nolint:gosec // why: test server 自己署名を信頼する。
			},
		},
	}
	return cursorapi.NewTextWriter(httpClient, apiKey), probe
}

func cursorNarrowStreamBody(text string) string {
	return "event: status\n" +
		`data: {"runId":"run-narrow","status":"RUNNING"}` + "\n\n" +
		"event: result\n" +
		`data: {"runId":"run-narrow","status":"FINISHED","text":"` + text + `"}` + "\n\n" +
		"event: done\n" +
		"data: {}\n\n"
}

func TestCursorTextWriter_deliversCreateAndStream_whenUpstreamSucceeds(t *testing.T) {
	// Given: dummy API key と、create → SSE 成功を通す upstream double
	const apiKey = "narrow-cursor-real-value"
	const fragment = "本日の IT ニュース原稿断片"
	writer, probe := newCursorTextWriterWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"agent":{"id":"bc-narrow","status":"ACTIVE"},"run":{"id":"run-narrow","status":"CREATING"}}`))
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cursorNarrowStreamBody(fragment)))
	})

	// When: Write する
	got, err := writer.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: upstream は POST(create) と GET(stream) を受け、Authorization に実値が届き、非空断片が返る
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != fragment {
		t.Fatalf("Write() = %q, want %q", got, fragment)
	}
	if probe.createMethod != http.MethodPost {
		t.Fatalf("create method = %q, want POST", probe.createMethod)
	}
	if probe.createAuth != "Bearer "+apiKey {
		t.Fatalf("create Authorization = %q, want %q", probe.createAuth, "Bearer "+apiKey)
	}
	if probe.streamMethod != http.MethodGet {
		t.Fatalf("stream method = %q, want GET", probe.streamMethod)
	}
	if !strings.Contains(probe.streamPath, "/v1/agents/bc-narrow/runs/run-narrow/stream") {
		t.Fatalf("stream path = %q", probe.streamPath)
	}
}

func TestCursorTextWriter_excludesDummySecretFromErrorMessage_whenUpstreamFails(t *testing.T) {
	// Given: dummy API key と、create が常に 401 を返す upstream double
	const apiKey = "narrow-cursor-must-not-leak-value"
	writer, probe := newCursorTextWriterWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})

	// When: Write する
	_, err := writer.Write(context.Background(), "narrow error message テスト")

	// Then: error は返るが dummy secret 実値は error message に出ない
	if err == nil {
		t.Fatal("Write() error = nil, want non-nil")
	}
	if probe.createMethod != http.MethodPost {
		t.Fatalf("create method = %q, want POST", probe.createMethod)
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("error message %q contains dummy secret value", err.Error())
	}
}
