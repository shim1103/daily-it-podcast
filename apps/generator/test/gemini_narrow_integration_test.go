// Scope: Narrow Integration
// 実物境界: gemini.SpeechSynthesizer が標準 *http.Client で送信する外向き HTTP request（test upstream server）
// Double: 本番 credential / Gemini 実 API は使わない。DialTLSContext で本番 host 宛先だけを test server へ redirect する。
// @require dummy API key を Adapter へ直接渡す。upstream は controllable な test server。
// @ensure upstream は POST を受け取り、x-goog-api-key header に実値が届く。
// @ensure 成功時 Synthesize は非空 WAV を返す。
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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

type geminiNarrowProbe struct {
	method string
	apiKey string
}

// newGeminiSynthesizerWithProxy は本番 host（generativelanguage.googleapis.com）への接続を test TLS server へ差し替えた SpeechSynthesizer を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure dummy API key は Adapter へ直接渡し、標準 *http.Client が x-goog-api-key header へ載せる。
func newGeminiSynthesizerWithProxy(t *testing.T, apiKey string, handler http.HandlerFunc) (*gemini.SpeechSynthesizer, *geminiNarrowProbe) {
	t.Helper()
	probe := &geminiNarrowProbe{}
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probe.method = r.Method
		probe.apiKey = r.Header.Get("x-goog-api-key")
		handler(w, r)
	}))
	t.Cleanup(upstream.Close)

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	return gemini.NewSpeechSynthesizer(httpClient, apiKey), probe
}

func isWAVFixture(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}

func TestGeminiSpeechSynthesizer_deliversPostWithAPIKeyHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: dummy API key と、成功応答を返す upstream double
	const apiKey = "narrow-gemini-real-value"
	synth, probe := newGeminiSynthesizerWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		writeIntegrationGeminiAudioResponse(t, w, minimalIntegrationGeminiPCM())
	})

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "本日の IT ニュースです。")

	// Then: upstream は POST を受け、x-goog-api-key に実値が届き、非空 WAV が返る
	if err != nil {
		t.Fatalf("Synthesize() error = %v, want nil", err)
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want %q", probe.method, http.MethodPost)
	}
	if probe.apiKey != apiKey {
		t.Fatalf("x-goog-api-key = %q, want %q", probe.apiKey, apiKey)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if !isWAVFixture(got.Content) {
		t.Fatalf("Content is not wav, head = % x", got.Content[:min(12, len(got.Content))])
	}
}

func TestGeminiSpeechSynthesizer_excludesDummySecretFromErrorMessage_whenUpstreamFails(t *testing.T) {
	// Given: dummy API key と、常に 400 を返す upstream double（非 retry で 1 回だけ）
	const apiKey = "narrow-gemini-must-not-leak-value"
	synth, probe := newGeminiSynthesizerWithProxy(t, apiKey, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"INVALID_ARGUMENT"}`))
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "narrow error message テスト")

	// Then: error は返るが、dummy secret 実値は error message に出ない
	if err == nil {
		t.Fatal("Synthesize() error = nil, want non-nil")
	}
	if probe.method != http.MethodPost {
		t.Fatalf("method = %q, want %q", probe.method, http.MethodPost)
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("error message %q contains dummy secret value", err.Error())
	}
}
