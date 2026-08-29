package gemini

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const geminiTestAPIKey = "gemini-test-real-value"

type proxyProbe struct {
	TargetURLs []string
	Methods    []string
	APIKeys    []string
	Bodies     []string
}

// why: Adapter は EndpointURL を定数として持つため、DialTLSContext で接続先だけを test server へ redirect する。
// why: synthesizer_edge_sociable_unit_test.go も参照する境界 I/O helper。Adapter 内分岐 case は
// fakeRoundTripper（境界 I/O なし）へ移行済みだが、edge file の見直しは別 issue の管轄。
func newSynthesizerWithProxy(t *testing.T, handler http.HandlerFunc) (*SpeechSynthesizer, *proxyProbe) {
	t.Helper()
	backoffNoWait := func(time.Duration) {}

	probe := &proxyProbe{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		probe.TargetURLs = append(probe.TargetURLs, r.URL.String())
		probe.Methods = append(probe.Methods, r.Method)
		probe.APIKeys = append(probe.APIKeys, r.Header.Get(geminiAPIKeyHeader))
		probe.Bodies = append(probe.Bodies, string(body))
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// why: test 用 TLS server の自己署名証明書を明示的に信頼する。
				return tls.Dial(network, server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
			},
		},
	}
	synth := newSpeechSynthesizerForTest(httpClient, geminiTestAPIKey, backoffNoWait)
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

// fakeClientCall は fakeRoundTripper が観測した request 1 件分。
type fakeClientCall struct {
	Method string
	URL    string
	APIKey string
	Body   []byte
}

// fakeRoundTripper は境界 I/O なしで http.RoundTripper を満たす Spy。
// 呼び出し順に responses を返し、各 request を記録する。
type fakeRoundTripper struct {
	responses []fakeClientResponse
	calls     []fakeClientCall
}

type fakeClientResponse struct {
	status int
	body   []byte
}

func (rt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	rt.calls = append(rt.calls, fakeClientCall{
		Method: req.Method,
		URL:    req.URL.String(),
		APIKey: req.Header.Get(geminiAPIKeyHeader),
		Body:   body,
	})
	index := len(rt.calls) - 1
	if index >= len(rt.responses) {
		return nil, fmt.Errorf("fakeRoundTripper: no response configured for call %d", index)
	}
	resp := rt.responses[index]
	rec := httptest.NewRecorder()
	rec.WriteHeader(resp.status)
	if resp.body != nil {
		_, _ = rec.Write(resp.body)
	}
	return rec.Result(), nil
}

func jsonBody(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return raw
}

func newFakeSynthesizer(responses ...fakeClientResponse) (*SpeechSynthesizer, *fakeRoundTripper) {
	rt := &fakeRoundTripper{responses: responses}
	synth := newSpeechSynthesizerForTest(&http.Client{Transport: rt}, "gemini-fake-key", func(time.Duration) {})
	return synth, rt
}

func TestSynthesize_wrapsTranscriptWithEnvelope_whenCallingProxy(t *testing.T) {

	// Given: 成功応答を返す Client Stub
	const transcript = "朗読する本文だけ"
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, audioInteractionResponse(minimalPCM())),
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), transcript)

	// Then: envelope + Transcript ラベル + 本文が input へ入る
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d", len(rt.calls))
	}
	var req map[string]any
	if err := json.Unmarshal(rt.calls[0].Body, &req); err != nil {
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

	// Given: 空本文。Client は呼ばれない想定なので response 設定は不要
	synth, rt := newFakeSynthesizer()

	// When: trim 後空の text を渡す
	_, err := synth.Synthesize(context.Background(), "  \t\n  ")

	// Then: Infrastructure Error。Client は呼ばない
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
	if len(rt.calls) != 0 {
		t.Fatalf("unexpected calls: %#v", rt.calls)
	}
}

func TestSynthesize_retriesTransientError_thenSucceeds(t *testing.T) {

	// Given: 1 回目 503、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
	)

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "retry テスト")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestSynthesize_returnsInfrastructureError_whenMaxAttemptsExceeded(t *testing.T) {

	// Given: 常に 503
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})}
	}
	synth, rt := newFakeSynthesizer(responses...)

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
	if len(rt.calls) != MaxAttempts {
		t.Fatalf("call count = %d, want %d", len(rt.calls), MaxAttempts)
	}
}

func TestSynthesize_doesNotRetry_whenStatusBadRequest(t *testing.T) {

	// Given: 400
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusBadRequest,
		body:   jsonBody(t, map[string]any{"error": "INVALID_ARGUMENT"}),
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "400 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesize_doesNotRetry_whenProhibitedContent(t *testing.T) {

	// Given: PROHIBITED_CONTENT
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, map[string]any{"error": map[string]any{"code": "PROHIBITED_CONTENT"}}),
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "禁止テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesize_retriesMissingAudio_whenStatusInternalError(t *testing.T) {

	// Given: 1 回目 500（audio 欠落）、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusInternalServerError, body: jsonBody(t, map[string]any{"error": "internal"})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
	)

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "500 retry")

	// Then: 2 回目で成功
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(got.Content) == 0 {
		t.Fatal("Content is empty")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}
