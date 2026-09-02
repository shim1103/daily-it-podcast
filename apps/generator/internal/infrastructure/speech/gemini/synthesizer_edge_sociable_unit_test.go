package gemini

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSynthesize_returnsInfrastructureError_whenOutputAudioMissingOnOK(t *testing.T) {

	// Given: HTTP 200 だが output_audio が無い（同種 retryable = Op "decode_pcm"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body:   jsonBody(t, map[string]any{"status": "ok"}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "audio 欠落")

	// Then: 同種 2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
	// Then: error 文言に body のトップレベルキー一覧が載り、fixture の "status" が含まれる
	if !strings.Contains(err.Error(), "top-level keys:") {
		t.Fatalf("Error() = %q, want it to contain %q", err.Error(), "top-level keys:")
	}
	if !strings.Contains(err.Error(), "status") {
		t.Fatalf("Error() = %q, want it to list top-level key %q", err.Error(), "status")
	}
}

func TestSynthesize_returnsInfrastructureError_whenResponseBodyInvalidJSON(t *testing.T) {

	// Given: 壊れた JSON（同種 retryable = Op "decode_pcm"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusOK, body: []byte(`not-json`)}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "decode 失敗")

	// Then: 同種 2 連続で打ち切り error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestSynthesize_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: nil client
	synth := NewSpeechSynthesizer(nil, "gemini-edge-key")

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "本文")

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSynthesize_retriesTooManyRequests_thenSucceeds(t *testing.T) {

	// Given: 1 回目 429、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusTooManyRequests, body: jsonBody(t, map[string]any{"error": "RESOURCE_EXHAUSTED"})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
	)

	// When: Synthesize する
	got, err := synth.Synthesize(context.Background(), "429 retry")

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

func TestSynthesize_doesNotRetry_whenStatusForbidden(t *testing.T) {

	// Given: 403
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusForbidden,
		body:   jsonBody(t, map[string]any{"error": "PERMISSION_DENIED"}),
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "403 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesize_returnsInfrastructureError_whenUnexpectedStatus(t *testing.T) {

	// Given: 404
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusNotFound,
		body:   jsonBody(t, map[string]any{"error": "NOT_FOUND"}),
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "404")

	// Then: retry しない
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesize_returnsInfrastructureError_whenBase64Invalid(t *testing.T) {

	// Given: 不正 base64（同種 retryable = Op "decode_pcm"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"steps": []map[string]any{
					{"content": []map[string]any{
						{"data": "!!!not-base64!!!"},
					}},
				},
			}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "bad b64")

	// Then: 同種 2 連続で打ち切り error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestSynthesize_returnsInfrastructureError_whenPCMLengthOdd(t *testing.T) {

	// Given: 奇数 byte の PCM
	oddPCM := []byte{0x00, 0x01, 0x02}
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body: jsonBody(t, map[string]any{
			"steps": []map[string]any{
				{"content": []map[string]any{
					{"data": base64.StdEncoding.EncodeToString(oddPCM)},
				}},
			},
		}),
	})

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "奇数 pcm")

	// Then: pcm 変換失敗（非 retry）
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesize_returnsInfrastructureError_whenReceiverNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var synth *SpeechSynthesizer

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "本文")

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *Error", err, infra)
	}
}

func TestSynthesize_returnsInfrastructureError_whenUpstreamDoFails(t *testing.T) {

	// Given: Client.Do が常に transport error を返す（同種 retryable = Op "do"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{err: fmt.Errorf("connection refused")}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.Synthesize(context.Background(), "network 失敗")

	// Then: 同種 2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}
