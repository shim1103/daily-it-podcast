package cursorapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// fakeClientCall は fakeRoundTripper が観測した request 1 件分。
type fakeClientCall struct {
	Method string
	URL    string
	Header http.Header
	Body   []byte
}

// fakeClientResponse は fakeRoundTripper が 1 回の呼び出しへ返す応答。
type fakeClientResponse struct {
	status int
	header http.Header
	body   string
	err    error
}

// fakeRoundTripper は境界 I/O なしで http.RoundTripper を満たす Spy。
// 呼び出し順に responses を返し、各 request を記録する。
type fakeRoundTripper struct {
	responses []fakeClientResponse
	calls     []fakeClientCall
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
		Header: req.Header.Clone(),
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
	for key, values := range resp.header {
		for _, v := range values {
			rec.Header().Add(key, v)
		}
	}
	rec.WriteHeader(resp.status)
	if resp.body != "" {
		_, _ = rec.WriteString(resp.body)
	}
	return rec.Result(), nil
}

func newFakeTextWriter(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper) {
	w, rt, _ := newFakeTextWriterWithSleepSpy(responses...)
	return w, rt
}

// sleepSpy は backoffSleepFn が観測した待ち時間を記録する。
type sleepSpy struct {
	waits []time.Duration
}

func newFakeTextWriterWithSleepSpy(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper, *sleepSpy) {
	rt := &fakeRoundTripper{responses: responses}
	spy := &sleepSpy{}
	w := newTextWriterForTest(&http.Client{Transport: rt}, "cursor-fake-key", func(_ context.Context, d time.Duration) {
		spy.waits = append(spy.waits, d)
	})
	return w, rt, spy
}

// createAgentBody は create 応答の fixture を組む。
func createAgentBody(agentID, runID string) string {
	raw, _ := json.Marshal(map[string]any{
		"agent": map[string]any{"id": agentID, "status": "ACTIVE"},
		"run":   map[string]any{"id": runID, "agentId": agentID, "status": "CREATING"},
	})
	return string(raw)
}

// sseStream は event/data 行を組み立てて text/event-stream 本文にする。
func sseStream(events ...sseEventFixture) string {
	var b strings.Builder
	for _, ev := range events {
		if ev.name != "" {
			b.WriteString("event: ")
			b.WriteString(ev.name)
			b.WriteString("\n")
		}
		b.WriteString("data: ")
		b.WriteString(ev.data)
		b.WriteString("\n\n")
	}
	return b.String()
}

type sseEventFixture struct {
	name string
	data string
}

// resultEvent は終端 result event の data JSON を組む。
func resultEvent(status, text string) sseEventFixture {
	raw, _ := json.Marshal(map[string]any{
		"runId":      "run-x",
		"status":     status,
		"text":       text,
		"durationMs": 1234,
	})
	return sseEventFixture{name: "result", data: string(raw)}
}

func successStreamResponse(text string) fakeClientResponse {
	return fakeClientResponse{
		status: http.StatusOK,
		header: http.Header{"Content-Type": {"text/event-stream"}},
		body: sseStream(
			sseEventFixture{name: "status", data: `{"runId":"run-x","status":"RUNNING"}`},
			sseEventFixture{name: "assistant", data: `{"text":"部分"}`},
			resultEvent("FINISHED", text),
			sseEventFixture{name: "done", data: `{}`},
		),
	}
}

func assertCursorInfraError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorapi.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "cursorapi:") {
		t.Fatalf("Error() = %q, want prefix cursorapi:", infra.Error())
	}
	if infra.Unwrap() == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func assertCursorInfraErrorOp(t *testing.T, err error, wantOp string) {
	t.Helper()
	assertCursorInfraError(t, err)
	var infra *Error
	_ = errors.As(err, &infra)
	if infra.Op != wantOp {
		t.Fatalf("Op = %q, want %q", infra.Op, wantOp)
	}
}

func TestWrite_returnsFragment_whenCreateThenStreamSucceeds(t *testing.T) {

	// Given: create 成功と、終端 result に非空 text を持つ SSE を返す Client Stub
	const fragment = "本日の IT ニュース原稿の断片"
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		successStreamResponse(fragment),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: 非空断片が返り、create は POST、stream は GET で叩かれる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != fragment {
		t.Fatalf("Write() = %q, want %q", got, fragment)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
	if rt.calls[0].Method != http.MethodPost {
		t.Fatalf("create method = %q, want POST", rt.calls[0].Method)
	}
	if rt.calls[0].URL != APIBaseURL+AgentsPath {
		t.Fatalf("create URL = %q, want %q", rt.calls[0].URL, APIBaseURL+AgentsPath)
	}
	if rt.calls[1].Method != http.MethodGet {
		t.Fatalf("stream method = %q, want GET", rt.calls[1].Method)
	}
	if !strings.Contains(rt.calls[1].URL, "/v1/agents/bc-1/runs/run-1/stream") {
		t.Fatalf("stream URL = %q, want .../v1/agents/bc-1/runs/run-1/stream", rt.calls[1].URL)
	}
	if got := rt.calls[0].Header.Get(AuthorizationHeader); got != BearerTokenPrefix+"cursor-fake-key" {
		t.Fatalf("Authorization = %q, want %q", got, BearerTokenPrefix+"cursor-fake-key")
	}
	var reqBody map[string]any
	if err := json.Unmarshal(rt.calls[0].Body, &reqBody); err != nil {
		t.Fatalf("decode create request: %v", err)
	}
	prompt, _ := reqBody["prompt"].(map[string]any)
	if prompt["text"] != "本文の要約から原稿を書いて" {
		t.Fatalf("prompt.text = %v", prompt["text"])
	}
	model, _ := reqBody["model"].(map[string]any)
	if model["id"] != ModelID {
		t.Fatalf("model.id = %v, want %q", model["id"], ModelID)
	}
	if _, hasRepos := reqBody["repos"]; hasRepos {
		t.Fatalf("create request must be no-repo, got repos: %v", reqBody["repos"])
	}
	if _, hasSource := reqBody["source"]; hasSource {
		t.Fatalf("create request must be no-repo, got source: %v", reqBody["source"])
	}
}

func TestWrite_returnsInfraError_whenBriefEmptyAfterTrim(t *testing.T) {

	// Given: trim 後空の brief。Client は呼ばれない想定なので response 設定は不要
	w, rt := newFakeTextWriter()

	// When: Write する
	got, err := w.Write(context.Background(), "  \t\n  ")

	// Then: validate_brief 系 Infra Error、断片空、Client は呼ばれない
	assertCursorInfraErrorOp(t, err, "validate_brief")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 0 {
		t.Fatalf("call count = %d, want 0", len(rt.calls))
	}
}

func TestWrite_retriesStreamOn429_untilMaxAttemptsThenInfraError(t *testing.T) {

	// Given: create 成功後、stream 取得が 429 を返し続ける
	responses := []fakeClientResponse{
		{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
	}
	for i := 0; i < MaxAttempts; i++ {
		responses = append(responses, fakeClientResponse{
			status: http.StatusTooManyRequests,
			header: http.Header{"Retry-After": {"1"}},
			body:   `{"error":"rate limited"}`,
		})
	}
	w, rt := newFakeTextWriter(responses...)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 上限到達で Infra Error、stream 取得は MaxAttempts 回
	assertCursorInfraErrorOp(t, err, "stream_status")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	streamCalls := len(rt.calls) - 1
	if streamCalls != MaxAttempts {
		t.Fatalf("stream call count = %d, want %d", streamCalls, MaxAttempts)
	}
}

func TestWrite_retriesStreamOn429_thenSucceeds(t *testing.T) {

	// Given: create 成功後、stream 取得が 1 回 429、2 回目で成功 SSE
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusTooManyRequests, body: `{"error":"rate limited"}`},
		successStreamResponse("再試行後の断片"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 2 回目の stream 取得で非空断片
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "再試行後の断片" {
		t.Fatalf("Write() = %q, want %q", got, "再試行後の断片")
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(rt.calls))
	}
}

func TestWrite_retriesStreamOnce_whenDoErrorThenSucceeds(t *testing.T) {

	// Given: create 成功後、stream 取得が Do error、2 回目で成功 SSE
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{err: fmt.Errorf("connection reset")},
		successStreamResponse("Do error 復帰後の断片"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: +1 即再試行で 2 回目成功、呼び出しは create + stream 2 回
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "Do error 復帰後の断片" {
		t.Fatalf("Write() = %q", got)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(rt.calls))
	}
}

func TestWrite_retriesStreamOnce_whenStatus5xxThenSucceeds(t *testing.T) {

	// Given: create 成功後、stream 取得が 503、2 回目で成功 SSE
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusServiceUnavailable, body: `{"error":"unavailable"}`},
		successStreamResponse("5xx 復帰後の断片"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: +1 即再試行で 2 回目成功
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "5xx 復帰後の断片" {
		t.Fatalf("Write() = %q", got)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(rt.calls))
	}
}

func TestWrite_doesNotRetryStreamTwice_whenStatus5xxPersists(t *testing.T) {

	// Given: create 成功後、stream 取得が 5xx を返し続ける
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusBadGateway, body: `{"error":"bad gateway"}`},
		fakeClientResponse{status: http.StatusBadGateway, body: `{"error":"bad gateway"}`},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 即再試行は 1 回だけ（stream 取得は 2 回）で Infra Error
	assertCursorInfraError(t, err)
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3 (create + stream x2)", len(rt.calls))
	}
}

func TestWrite_doesNotRetryCreate_whenStatus5xx(t *testing.T) {

	// Given: create（POST /v1/agents）が 5xx を返す
	w, rt := newFakeTextWriter(fakeClientResponse{
		status: http.StatusInternalServerError,
		body:   `{"error":"internal"}`,
	})

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 非 idempotent なので再試行しない（呼び出し 1 回）で Infra Error
	assertCursorInfraErrorOp(t, err, "create_status")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenResultTextEmpty(t *testing.T) {

	// Given: create 成功後、終端 result の text が空
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{
			status: http.StatusOK,
			header: http.Header{"Content-Type": {"text/event-stream"}},
			body:   sseStream(resultEvent("FINISHED", "   "), sseEventFixture{name: "done", data: `{}`}),
		},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 非 retry の Infra Error（stream 取得は 1 回だけ）
	assertCursorInfraErrorOp(t, err, "empty_text")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenRunTerminatedWithError(t *testing.T) {

	// Given: create 成功後、終端 result の status が ERROR
	w, _ := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{
			status: http.StatusOK,
			header: http.Header{"Content-Type": {"text/event-stream"}},
			body:   sseStream(resultEvent("ERROR", ""), sseEventFixture{name: "done", data: `{}`}),
		},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: run 終端 error は非 retry の Infra Error
	assertCursorInfraErrorOp(t, err, "run_status")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
}

func TestWrite_excludesAPIKeyFromErrorMessage_whenCreateFails(t *testing.T) {

	// Given: create が 401 を返す。fake key の実値を error に出してはならない
	const apiKey = "cursor-secret-must-not-leak"
	rt := &fakeRoundTripper{responses: []fakeClientResponse{
		{status: http.StatusUnauthorized, body: `{"error":"unauthorized"}`},
	}}
	w := newTextWriterForTest(&http.Client{Transport: rt}, apiKey, func(context.Context, time.Duration) {})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて")

	// Then: error は返るが apiKey 実値は message に出ない
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), apiKey) {
		t.Fatalf("error message %q contains api key value", err.Error())
	}
}

func TestWrite_doesNotRetryCreate_whenClientErrorStatus(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {

			// Given: create が client error status を返す
			w, rt := newFakeTextWriter(fakeClientResponse{
				status: status,
				body:   `{"error":"denied"}`,
			})

			// When: Write する
			got, err := w.Write(context.Background(), "原稿を書いて")

			// Then: 非 retry の Infra Error、断片空、呼び出しは 1 回だけ
			assertCursorInfraErrorOp(t, err, "create_status")
			if got != "" {
				t.Fatalf("fragment = %q, want empty", got)
			}
			if len(rt.calls) != 1 {
				t.Fatalf("call count = %d, want 1", len(rt.calls))
			}
		})
	}
}

func TestWrite_doesNotRetryStream_whenClientErrorStatus(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {

			// Given: create 成功後、stream 取得が client error status を返す
			w, rt := newFakeTextWriter(
				fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
				fakeClientResponse{status: status, body: `{"error":"denied"}`},
			)

			// When: Write する
			got, err := w.Write(context.Background(), "原稿を書いて")

			// Then: 非 retry の Infra Error、断片空、呼び出しは create + stream の 2 回だけ
			assertCursorInfraErrorOp(t, err, "stream_status")
			if got != "" {
				t.Fatalf("fragment = %q, want empty", got)
			}
			if len(rt.calls) != 2 {
				t.Fatalf("call count = %d, want 2", len(rt.calls))
			}
		})
	}
}

func TestWrite_doesNotReStream_whenBodyEndsBeforeResultEvent(t *testing.T) {

	// Given: create 成功後、stream は 200 だが result event 前に body 終端
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{
			status: http.StatusOK,
			header: http.Header{"Content-Type": {"text/event-stream"}},
			body: sseStream(
				sseEventFixture{name: "status", data: `{"runId":"run-1","status":"RUNNING"}`},
				sseEventFixture{name: "assistant", data: `{"text":"途中"}`},
			),
		},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: parse_sse Infra Error。再 stream しない（呼び出しは create + stream の 2 回だけ）
	assertCursorInfraErrorOp(t, err, "parse_sse")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestCtxSleep_returnsEarly_whenContextAlreadyCancelled(t *testing.T) {

	// Given: 既に cancel 済みの ctx と、実測できないほど長い待ち時間
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When: ctxSleep を呼ぶ
	start := time.Now()
	ctxSleep(ctx, time.Hour)

	// Then: timer を待たずに即戻る
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("ctxSleep blocked for %v, want near-immediate return", elapsed)
	}
}

func TestWrite_clampsRetryAfter_whenHeaderValueExceedsMax(t *testing.T) {

	// Given: create 成功後、stream が過大な Retry-After 付き 429、その後成功 SSE
	w, _, spy := newFakeTextWriterWithSleepSpy(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{
			status: http.StatusTooManyRequests,
			header: http.Header{"Retry-After": {"999999"}},
			body:   `{"error":"rate limited"}`,
		},
		successStreamResponse("クランプ後の断片"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 成功断片が返り、待ち時間は MaxRetryAfter でクランプされる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "クランプ後の断片" {
		t.Fatalf("Write() = %q", got)
	}
	if len(spy.waits) != 1 {
		t.Fatalf("sleep count = %d, want 1", len(spy.waits))
	}
	if spy.waits[0] != MaxRetryAfter {
		t.Fatalf("wait = %v, want %v (clamped)", spy.waits[0], MaxRetryAfter)
	}
}
