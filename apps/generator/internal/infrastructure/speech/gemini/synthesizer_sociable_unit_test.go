package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

type proxyProbe struct {
	TargetURLs []string
	Methods    []string
	APIKeys    []string
	Bodies     []string
}

func newSynthesizerWithProxy(t *testing.T, handler http.HandlerFunc) (*SpeechSynthesizer, *proxyProbe) {
	t.Helper()
	backoffNoWait := func(time.Duration) {}

	probe := &proxyProbe{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		probe.TargetURLs = append(probe.TargetURLs, r.Header.Get("X-AS-Target-URL"))
		probe.Methods = append(probe.Methods, r.Header.Get("X-AS-Method"))
		probe.APIKeys = append(probe.APIKeys, r.Header.Get("X-AS-Inject-Header-x-goog-api-key"))
		probe.Bodies = append(probe.Bodies, string(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	synth := newSpeechSynthesizerForTest(&agentsecrets.Client{
		HTTP:     server.Client(),
		ProxyURL: server.URL,
	}, backoffNoWait)
	return synth, probe
}

func minimalPCM() []byte {
	// why: 24 kHz / 16-bit / mono。短い無音でも WAV wrap できる長さにする。
	const sampleCount = 2400
	pcm := make([]byte, sampleCount*2)
	return pcm
}

func audioInteractionResponse(pcm []byte) map[string]any {
	return map[string]any{
		"output_audio": map[string]any{
			"data": base64.StdEncoding.EncodeToString(pcm),
		},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

func isWAV(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}

func TestSynthesize_returnsNonEmptyWAV_whenProxyReturnsPCM(t *testing.T) {
	t.Parallel()

	// Given: proxy が PCM を返す
	synth, _ := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, audioInteractionResponse(minimalPCM()))
	})

	// When: 本文を渡して Synthesize する
	got, err := synth.Synthesize(context.Background(), "本日の IT ニュースです。")

	// Then: 非空 WAV が返る
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if !isWAV(got.Content) {
		t.Fatalf("Content is not wav, head = % x", got.Content[:min(12, len(got.Content))])
	}
}

func TestSynthesize_injectsGeminiAPIKeyName_whenCallingProxy(t *testing.T) {
	t.Parallel()

	// Given: 成功する proxy double
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, audioInteractionResponse(minimalPCM()))
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "テスト本文")

	// Then: x-goog-api-key にキー名だけが渡る
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(probe.APIKeys) != 1 {
		t.Fatalf("api key headers = %#v", probe.APIKeys)
	}
	if probe.APIKeys[0] != secretnames.GeminiAPIKeyName {
		t.Fatalf("api key = %q, want %q", probe.APIKeys[0], secretnames.GeminiAPIKeyName)
	}
	if len(probe.Methods) != 1 || probe.Methods[0] != http.MethodPost {
		t.Fatalf("methods = %#v", probe.Methods)
	}
	if len(probe.TargetURLs) != 1 || probe.TargetURLs[0] != EndpointURL {
		t.Fatalf("target URLs = %#v", probe.TargetURLs)
	}
}

func TestSynthesize_wrapsTranscriptWithEnvelope_whenCallingProxy(t *testing.T) {
	t.Parallel()

	// Given: リクエスト body を観測できる proxy double
	const transcript = "朗読する本文だけ"
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, audioInteractionResponse(minimalPCM()))
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), transcript)

	// Then: envelope + Transcript ラベル + 本文が input へ入る
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(probe.Bodies) != 1 {
		t.Fatalf("bodies = %d", len(probe.Bodies))
	}
	var req map[string]any
	if err := json.Unmarshal([]byte(probe.Bodies[0]), &req); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	input, _ := req["input"].(string)
	if !strings.Contains(input, EnvelopePreamble) {
		t.Fatalf("input missing preamble: %q", input)
	}
	if !strings.Contains(input, TranscriptLabel) {
		t.Fatalf("input missing transcript label: %q", input)
	}
	if !strings.Contains(input, transcript) {
		t.Fatalf("input missing transcript: %q", input)
	}
	genCfg, _ := req["generation_config"].(map[string]any)
	speechCfg, _ := genCfg["speech_config"].([]any)
	if len(speechCfg) != 1 {
		t.Fatalf("speech_config = %#v", speechCfg)
	}
	voiceCfg, _ := speechCfg[0].(map[string]any)
	if voiceCfg["voice"] != VoiceName {
		t.Fatalf("voice = %v, want %q", voiceCfg["voice"], VoiceName)
	}
	if req["model"] != ModelID {
		t.Fatalf("model = %v, want %q", req["model"], ModelID)
	}
}

func TestSynthesize_returnsInfrastructureError_whenTextEmptyAfterTrim(t *testing.T) {
	t.Parallel()

	// Given: 空本文
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("proxy must not be called")
	})

	// When: trim 後空の text を渡す
	_, err := synth.Synthesize(context.Background(), "  \t\n  ")

	// Then: Infrastructure Error。HTTP は呼ばない
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gemini.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "gemini:") {
		t.Fatalf("Error() = %q, want prefix gemini:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
	if len(probe.TargetURLs) != 0 {
		t.Fatalf("unexpected requests: %#v", probe.TargetURLs)
	}
}

func TestSynthesize_retriesTransientError_thenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目 503、2 回目成功
	var calls atomic.Int32
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"error": "UNAVAILABLE"})
			return
		}
		writeJSON(t, w, http.StatusOK, audioInteractionResponse(minimalPCM()))
	})

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "retry テスト")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(probe.TargetURLs) != 2 {
		t.Fatalf("request count = %d, want 2", len(probe.TargetURLs))
	}
}

func TestSynthesize_returnsInfrastructureError_whenMaxAttemptsExceeded(t *testing.T) {
	t.Parallel()

	// Given: 常に 503
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusServiceUnavailable, map[string]any{"error": "UNAVAILABLE"})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "打ち切りテスト")

	// Then: MaxAttempts 回だけ試し Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gemini.Error", err, err)
	}
	if len(probe.TargetURLs) != MaxAttempts {
		t.Fatalf("request count = %d, want %d", len(probe.TargetURLs), MaxAttempts)
	}
}

func TestSynthesize_doesNotRetry_whenStatusBadRequest(t *testing.T) {
	t.Parallel()

	// Given: 400
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusBadRequest, map[string]any{"error": "INVALID_ARGUMENT"})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "400 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != 1 {
		t.Fatalf("request count = %d, want 1", len(probe.TargetURLs))
	}
}

func TestSynthesize_doesNotRetry_whenProhibitedContent(t *testing.T) {
	t.Parallel()

	// Given: PROHIBITED_CONTENT
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"error": map[string]any{"code": "PROHIBITED_CONTENT"},
		})
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "禁止テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(probe.TargetURLs) != 1 {
		t.Fatalf("request count = %d, want 1", len(probe.TargetURLs))
	}
}

func TestSynthesize_retriesMissingAudio_whenStatusInternalError(t *testing.T) {
	t.Parallel()

	// Given: 1 回目 500（audio 欠落）、2 回目成功
	var calls atomic.Int32
	synth, probe := newSynthesizerWithProxy(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			writeJSON(t, w, http.StatusInternalServerError, map[string]any{"error": "internal"})
			return
		}
		writeJSON(t, w, http.StatusOK, audioInteractionResponse(minimalPCM()))
	})

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "500 retry")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(probe.TargetURLs) != 2 {
		t.Fatalf("request count = %d, want 2", len(probe.TargetURLs))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
