// Scope: Narrow Integration
// 実物境界: geminiapi.TextWriter が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 credential / Gemini 実 API は使わない。DialTLSContext で本番 host（generativelanguage.googleapis.com）宛先だけを test server へ redirect する。
// @require dummy API key を Adapter へ直接渡す。upstream は controllable な test server。
// @ensure upstream は generateContent の POST を受け取り、x-goog-api-key header に実値が届く。
// @ensure 成功時 Write は非空断片を返す。
// @invariant dummy secret 実値は error message へ出ない。key は URL query に載らない。
package test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/geminiapi"
)

type geminiManuscriptNarrowProbe struct {
	method   string
	apiKey   string
	rawQuery string
}

// newGeminiTextWriterWithProxy は本番 host（generativelanguage.googleapis.com）への接続を test TLS server へ差し替えた TextWriter を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure dummy API key は Adapter へ直接渡し、標準 *http.Client が x-goog-api-key header へ載せる。
func newGeminiTextWriterWithProxy(t *testing.T, apiKey string, handler http.HandlerFunc) (*geminiapi.TextWriter, *geminiManuscriptNarrowProbe) {
	t.Helper()
	probe := &geminiManuscriptNarrowProbe{}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probe.method = r.Method
		probe.apiKey = r.Header.Get("x-goog-api-key")
		probe.rawQuery = r.URL.RawQuery
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
	return geminiapi.NewTextWriter(httpClient, apiKey), probe
}

// geminiManuscriptNarrowBody は generateContent 成功応答（candidates + finishReason: STOP + 非空 text）の fixture を組む。
func geminiManuscriptNarrowBody(text string) string {
	raw, _ := json.Marshal(map[string]any{
		"candidates": []map[string]any{
			{
				"content":      map[string]any{"parts": []map[string]any{{"text": text}}, "role": "model"},
				"finishReason": "STOP",
			},
		},
		"usageMetadata": map[string]any{"promptTokenCount": 12, "candidatesTokenCount": 34},
	})
	return string(raw)
}

func TestGeminiTextWriter_deliversPostWithAPIKeyHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: dummy API key と、generateContent 成功応答を返す upstream double
	const apiKey = "narrow-gemini-manuscript-real-value"
	const fragment = "本日の IT ニュース原稿の断片"
	writer, probe := newGeminiTextWriterWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(geminiManuscriptNarrowBody(fragment)))
	})

	// When: Write する
	got, err := writer.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: upstream は POST を受け、x-goog-api-key に実値が届き、非空断片が返る。key は URL query に出ない
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != fragment {
		t.Fatalf("Write() = %q, want %q", got, fragment)
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want %q", probe.method, http.MethodPost)
	}
	if probe.apiKey != apiKey {
		t.Fatalf("x-goog-api-key = %q, want %q", probe.apiKey, apiKey)
	}
	if strings.Contains(probe.rawQuery, apiKey) || strings.Contains(probe.rawQuery, "key=") {
		t.Fatalf("URL RawQuery = %q, want no api key", probe.rawQuery)
	}
}

func TestGeminiTextWriter_excludesDummySecretFromErrorMessage_whenUpstreamFails(t *testing.T) {
	// Given: dummy API key と、常に 400 を返す upstream double（非 retry で 1 回だけ）
	const apiKey = "narrow-gemini-manuscript-must-not-leak-value"
	writer, probe := newGeminiTextWriterWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"INVALID_ARGUMENT"}`))
	})

	// When: Write する
	_, err := writer.Write(context.Background(), "narrow error message テスト")

	// Then: error は返るが、dummy secret 実値は error message に出ない
	if err == nil {
		t.Fatal("Write() error = nil, want non-nil")
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want %q", probe.method, http.MethodPost)
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("error message %q contains dummy secret value", err.Error())
	}
}
