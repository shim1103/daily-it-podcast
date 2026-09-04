package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSynthesizeOne_returnsInfrastructureError_whenTextEmptyAfterTrim(t *testing.T) {

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

func TestSynthesizeOne_retriesTransientError_thenSucceeds(t *testing.T) {

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

func TestSynthesizeOne_returnsInfrastructureError_whenSameRetryableOpRepeatsTwice(t *testing.T) {

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

func TestSynthesizeOne_retriesUpToMaxAttempts_whenRetryableOpAlternates(t *testing.T) {

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

func TestSynthesizeOne_retriesMissingAudio_whenStatusInternalError(t *testing.T) {

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

func TestSynthesizeOne_retriesTooManyRequests_thenSucceeds(t *testing.T) {

	// Given: 1 回目 429、2 回目成功
	synth, rt := newFakeSynthesizer(
		fakeClientResponse{status: http.StatusTooManyRequests, body: jsonBody(t, map[string]any{"error": "RESOURCE_EXHAUSTED"})},
		fakeClientResponse{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
	)

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "429 retry")

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

func TestSynthesizeOne_doesNotRetry_whenStatusBadRequest(t *testing.T) {

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

func TestSynthesizeOne_doesNotRetry_whenStatusForbidden(t *testing.T) {

	// Given: 403
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusForbidden,
		body:   jsonBody(t, map[string]any{"error": "PERMISSION_DENIED"}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "403 テスト")

	// Then: 1 回だけ
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_returnsInfrastructureError_whenUnexpectedStatus(t *testing.T) {

	// Given: 404
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusNotFound,
		body:   jsonBody(t, map[string]any{"error": "NOT_FOUND"}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "404")

	// Then: retry しない
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestSynthesizeOne_doesNotRetry_whenProhibitedContent(t *testing.T) {

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

func TestSynthesizeOne_returnsInfrastructureError_whenUpstreamDoFails(t *testing.T) {

	// Given: Client.Do が常に transport error を返す（同種 retryable = Op "do"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{err: fmt.Errorf("connection refused")}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "network 失敗")

	// Then: 同種 2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}
