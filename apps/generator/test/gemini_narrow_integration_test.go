// Scope: Narrow Integration
// 実物境界: gemini.SpeechSynthesizer が processenv.Client 経由で送信する外向き HTTP request（test upstream server）
// Double: BindingResolver は Composition と同型の in-memory map。本番 credential / Gemini 実 API は使わない。
// @require dummy process environment（t.Setenv）に secret 実値をセットする。upstream は controllable な test server。
// @ensure upstream は POST を受け取り、x-goog-api-key header に実値が届く。
// @ensure 成功時 Synthesize は非空 WAV を返す。
// @invariant dummy secret 実値は error message へ出ない。
package test

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

const geminiNarrowTestSecretName = "NARROW_GEMINI_API_KEY"

// newGeminiSynthesizerWithProxy は EndpointURL への接続を test server へ redirect した SpeechSynthesizer を返す。
//
// @require handler は upstream request を観測・応答する。
// @ensure dummy secret 実値は t.Setenv だけに存在し、synthesizer は processenv.Client 経由で実注入する。
func newGeminiSynthesizerWithProxy(t *testing.T, secretValue string, handler http.HandlerFunc) *gemini.SpeechSynthesizer {
	t.Helper()

	upstream := httptest.NewTLSServer(handler)
	t.Cleanup(upstream.Close)
	t.Setenv(geminiNarrowTestSecretName, secretValue)

	apiKeySecret := secrettransport.NewSecretRef()
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, upstream.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	client := processenv.NewClient(narrowBindings{apiKeySecret: geminiNarrowTestSecretName}, httpClient, nil)
	return gemini.NewSpeechSynthesizer(client, apiKeySecret)
}

func minimalGeminiPCM() []byte {
	// why: 24 kHz / 16-bit / mono。短い無音でも WAV wrap できる長さにする。
	const sampleCount = 2400
	return make([]byte, sampleCount*2)
}

func writeGeminiAudioResponse(t *testing.T, w http.ResponseWriter, pcm []byte) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body := `{"output_audio":{"data":"` + base64.StdEncoding.EncodeToString(pcm) + `"}}`
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func isWAVFixture(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}

func TestGeminiSpeechSynthesizer_deliversPostWithAPIKeyHeader_whenUpstreamSucceeds(t *testing.T) {
	// Given: dummy secret 実値と、成功応答を返す upstream double
	const secretValue = "narrow-gemini-real-value"
	var gotMethod string
	var gotAPIKey string
	synth := newGeminiSynthesizerWithProxy(t, secretValue, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAPIKey = r.Header.Get("x-goog-api-key")
		writeGeminiAudioResponse(t, w, minimalGeminiPCM())
	})

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "本日の IT ニュースです。")

	// Then: upstream は POST を受け、x-goog-api-key に実値が届き、非空 WAV が返る
	if err != nil {
		t.Fatalf("Synthesize() error = %v, want nil", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotAPIKey != secretValue {
		t.Fatalf("x-goog-api-key = %q, want %q", gotAPIKey, secretValue)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if !isWAVFixture(got.Content) {
		t.Fatalf("Content is not wav, head = % x", got.Content[:min(12, len(got.Content))])
	}
}

func TestGeminiSpeechSynthesizer_excludesDummySecretFromErrorMessage_whenUpstreamFails(t *testing.T) {
	// Given: dummy secret 実値と、常に 400 を返す upstream double（非 retry で 1 回だけ）
	const secretValue = "narrow-gemini-must-not-leak-value"
	synth := newGeminiSynthesizerWithProxy(t, secretValue, func(w http.ResponseWriter, r *http.Request) {
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
	if strings.Contains(err.Error(), secretValue) {
		t.Fatalf("error message %q contains dummy secret value", err.Error())
	}
}
