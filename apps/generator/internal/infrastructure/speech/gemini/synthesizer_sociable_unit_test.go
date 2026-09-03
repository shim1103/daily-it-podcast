package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

func minimalPCM() []byte {
	// why: 24 kHz / 16-bit / mono。最小尺閾値（minPCMBytes = 0.5s）を超える長さにする。
	//      これ未満だと decodePCM が「極小 PCM」として retryable な失敗に落とす。
	const sampleCount = 24000 // 1.0s 相当（24000 * 2 = 48000 bytes > minPCMBytes 24000）
	pcm := make([]byte, sampleCount*2)
	return pcm
}

func audioInteractionResponse(pcm []byte) map[string]any {
	return map[string]any{
		"status": "completed",
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(pcm)},
			}},
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
	err    error
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
		Body:   body,
	})
	index := len(rt.calls) - 1
	if index >= len(rt.responses) {
		return nil, fmt.Errorf("fakeRoundTripper: no response configured for call %d", index)
	}
	resp := rt.responses[index]
	if resp.err != nil {
		return nil, resp.err
	}
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

// synthTestOne は 1 セグメント分の retry ループ（synthesizeOne）を MaxAttempts 上限で叩く test helper。
// 単一セグメントの loop 挙動を検証する既存 test 用。合計予算・入口ガードの検証は SynthesizeAll の test が持つ。
func (s *SpeechSynthesizer) synthTestOne(ctx context.Context, text string) (models.SpeechAudio, error) {
	audio, _, err := s.synthesizeOne(ctx, text, MaxAttempts)
	return audio, err
}

func TestSynthesize_wrapsTranscriptWithEnvelope_whenCallingProxy(t *testing.T) {

	// Given: 成功応答を返す Client Stub
	const transcript = "朗読する本文だけ"
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, audioInteractionResponse(minimalPCM())),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), transcript)

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
	_, err := synth.synthTestOne(context.Background(), "  \t\n  ")

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
	got, err := synth.synthTestOne(context.Background(), "retry テスト")

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

func TestSynthesize_returnsInfrastructureError_whenSameRetryableOpRepeatsTwice(t *testing.T) {

	// Given: 常に 503（同種 retryable error = Op "http_status"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "打ち切りテスト")

	// Then: 同種 error 2 連続で打ち切り、call は 2 回で Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gemini.Error", err, err)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestSynthesize_retriesUpToMaxAttempts_whenRetryableOpAlternates(t *testing.T) {

	// Given: 503（http_status）と transport error（do）が交互。Op が変わるので連続打ち切りに掛からない
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
		fakeClientResponse{err: fmt.Errorf("connection reset")},
		fakeClientResponse{status: http.StatusServiceUnavailable, body: jsonBody(t, map[string]any{"error": "UNAVAILABLE"})},
	)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "交互リトライ")

	// Then: MaxAttempts 回まで回って Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != MaxAttempts {
		t.Fatalf("call count = %d, want %d（Op 交互は打ち切らない）", len(rt.calls), MaxAttempts)
	}
}

func TestSynthesize_doesNotRetry_whenStatusBadRequest(t *testing.T) {

	// Given: 400
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusBadRequest,
		body:   jsonBody(t, map[string]any{"error": "INVALID_ARGUMENT"}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "400 テスト")

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
	_, err := synth.synthTestOne(context.Background(), "禁止テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesize_extractsAudioFromStepsContentData_whenRealResponseShape(t *testing.T) {

	// Given: 実 Interactions API そっくりの応答（steps[].content[].data に audio base64）
	pcm := minimalPCM()
	body := map[string]any{
		"id":           "v1_real_shape",
		"object":       "interaction",
		"model":        ModelID,
		"status":       "completed",
		"service_tier": "standard",
		"created":      "2026-09-02T08:36:33Z",
		"updated":      "2026-09-02T08:36:33Z",
		"usage": map[string]any{
			"total_tokens":        125,
			"total_output_tokens": 99,
			"output_tokens_by_modality": []map[string]any{
				{"modality": "audio", "tokens": 99},
			},
		},
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(pcm)},
			}},
		},
	}
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, body),
	})

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "実応答形テスト")

	// Then: steps 構造から audio を取り出し非空 WAV を返す
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if !isWAV(got.Content) {
		t.Fatalf("Content is not WAV: %d bytes", len(got.Content))
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
	got, err := synth.synthTestOne(context.Background(), "500 retry")

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
