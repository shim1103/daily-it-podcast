package geminiapi

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

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
)

// Scope: Sociable Unit
// 実物: geminiapi.TextWriter（Gemini generateContent Adapter）
// Double: http.RoundTripper の Spy（fakeRoundTripper）。backoffSleepFn は待ちを観測する sleepSpy。
//
// retry / error 方針: 成功応答で非空断片 / client.Do error・5xx を +1 即再試行 / 429 を MaxAttempts まで backoff /
// 401・403・その他 4xx は非 retry / finishReason≠STOP・空 text は非 retry / secret は error へ出さず header に載る /
// geminiapi は番兵 port.ErrSourceExhausted を出さない。

const fakeAPIKey = "gemini-fake-key-must-not-leak"

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
	// bodyReadErr を設定すると、応答本文の読み取り途中で此の error を返す（transient network 切断の模倣）。
	bodyReadErr error
}

// errAfterReader は先頭 bytes を返した後 err を返す io.ReadCloser。body read 途中断の模倣に使う。
type errAfterReader struct {
	data []byte
	err  error
}

func (r *errAfterReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, r.err
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func (r *errAfterReader) Close() error { return nil }

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
	result := rec.Result()
	if resp.bodyReadErr != nil {
		result.Body = &errAfterReader{data: []byte(resp.body), err: resp.bodyReadErr}
	}
	return result, nil
}

// sleepSpy は backoffSleepFn が観測した待ち時間を記録する。
type sleepSpy struct {
	waits []time.Duration
}

func newFakeTextWriter(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper) {
	w, rt, _ := newFakeTextWriterWithSleepSpy(responses...)
	return w, rt
}

func newFakeTextWriterWithSleepSpy(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper, *sleepSpy) {
	rt := &fakeRoundTripper{responses: responses}
	spy := &sleepSpy{}
	w := newTextWriter(&http.Client{Transport: rt}, fakeAPIKey, func(_ context.Context, d time.Duration) {
		spy.waits = append(spy.waits, d)
	})
	return w, rt, spy
}

// generateContentBody は generateContent 成功応答の fixture を組む。
// why: request は candidateCount 未指定なので実 API は 1 候補のみ返す。fixture も candidates を 1 要素に固定し、
//
//	parts は複数（連結対象）を許す。
func generateContentBody(finishReason string, texts ...string) string {
	parts := make([]map[string]any, 0, len(texts))
	for _, tx := range texts {
		parts = append(parts, map[string]any{"text": tx})
	}
	raw, _ := json.Marshal(map[string]any{
		"candidates": []map[string]any{
			{
				"content":      map[string]any{"parts": parts, "role": "model"},
				"finishReason": finishReason,
			},
		},
		"usageMetadata": map[string]any{"promptTokenCount": 10, "candidatesTokenCount": 20},
	})
	return string(raw)
}

func successResponse(finishReason string, texts ...string) fakeClientResponse {
	return fakeClientResponse{
		status: http.StatusOK,
		header: http.Header{"Content-Type": {"application/json"}},
		body:   generateContentBody(finishReason, texts...),
	}
}

func assertGeminiInfraError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "geminiapi:") {
		t.Fatalf("Error() = %q, want prefix geminiapi:", infra.Error())
	}
	if infra.Unwrap() == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func assertGeminiInfraErrorOp(t *testing.T, err error, wantOp string) {
	t.Helper()
	assertGeminiInfraError(t, err)
	var infra *adaptererror.Error
	_ = errors.As(err, &infra)
	if infra.Op != wantOp {
		t.Fatalf("Op = %q, want %q", infra.Op, wantOp)
	}
}

func TestWrite_returnsFragment_whenGenerateContentSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 候補・finishReason STOP・非空 text を 2 part 返す成功応答（parts を連結する前提）
	const brief = "本文の要約から原稿を書いて"
	w, rt := newFakeTextWriter(successResponse("STOP", "前半の原稿。", "後半の原稿。"))

	// When: Write する
	got, err := w.Write(context.Background(), brief)

	// Then: part の text を連結した非空断片が返る
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "前半の原稿。後半の原稿。" {
		t.Fatalf("Write() = %q, want %q", got, "前半の原稿。後半の原稿。")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
	if rt.calls[0].Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", rt.calls[0].Method)
	}
	wantURL := fmt.Sprintf(EndpointURLTemplate, ModelID)
	if rt.calls[0].URL != wantURL {
		t.Fatalf("URL = %q, want %q", rt.calls[0].URL, wantURL)
	}
	var reqBody struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(rt.calls[0].Body, &reqBody); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if len(reqBody.Contents) != 1 || len(reqBody.Contents[0].Parts) != 1 {
		t.Fatalf("request body shape = %+v", reqBody)
	}
	if reqBody.Contents[0].Parts[0].Text != brief {
		t.Fatalf("contents[0].parts[0].text = %q, want %q", reqBody.Contents[0].Parts[0].Text, brief)
	}
}

func TestWrite_retriesOnce_whenDoErrorThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目が client.Do error、2 回目で成功応答
	w, rt := newFakeTextWriter(
		fakeClientResponse{err: fmt.Errorf("connection reset")},
		successResponse("STOP", "Do error 復帰後の原稿。"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: +1 即再試行で 2 回目成功（calls == 2）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "Do error 復帰後の原稿。" {
		t.Fatalf("Write() = %q", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_retriesOnce_whenStatus5xxThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目が 503、2 回目で成功応答
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusServiceUnavailable, body: `{"error":"unavailable"}`},
		successResponse("STOP", "5xx 復帰後の原稿。"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: +1 即再試行で 2 回目成功（calls == 2）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "5xx 復帰後の原稿。" {
		t.Fatalf("Write() = %q", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_doesNotRetryTwice_whenStatus5xxPersists(t *testing.T) {
	t.Parallel()

	// Given: 5xx が続く
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusBadGateway, body: `{"error":"bad gateway"}`},
		fakeClientResponse{status: http.StatusBadGateway, body: `{"error":"bad gateway"}`},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 即再試行は 1 回だけ（calls == 2）で *adaptererror.Error
	assertGeminiInfraErrorOp(t, err, "http_status")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_retriesOnceThenSucceeds_whenBodyReadIsInterrupted(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は body 読み取り途中で切断、2 回目は成功応答
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: `{"candi`, bodyReadErr: errors.New("connection reset by peer")},
		successResponse("STOP", "再接続後の原稿。"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: read_body 断は一過性として 1 回だけ再試行し、2 回目で成功する（calls == 2）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "再接続後の原稿。" {
		t.Fatalf("Write() = %q", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_doesNotRetryTwice_whenBodyReadInterruptionPersists(t *testing.T) {
	t.Parallel()

	// Given: body 読み取り途中断が 2 回続く
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: `{"candi`, bodyReadErr: errors.New("connection reset by peer")},
		fakeClientResponse{status: http.StatusOK, body: `{"candi`, bodyReadErr: errors.New("connection reset by peer")},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 即再試行は 1 回だけ（calls == 2）で read_body の Infrastructure Error
	assertGeminiInfraErrorOp(t, err, "read_body")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_retriesOn429_untilMaxAttemptsThenInfraError(t *testing.T) {
	t.Parallel()

	// Given: 429 を MaxAttempts 回返し続ける
	responses := make([]fakeClientResponse, 0, MaxAttempts)
	for i := 0; i < MaxAttempts; i++ {
		responses = append(responses, fakeClientResponse{
			status: http.StatusTooManyRequests,
			body:   `{"error":"rate limited"}`,
		})
	}
	w, rt, spy := newFakeTextWriterWithSleepSpy(responses...)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 上限到達で *adaptererror.Error、待ち観測は MaxAttempts-1 回、call は MaxAttempts 回
	assertGeminiInfraErrorOp(t, err, "http_status")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != MaxAttempts {
		t.Fatalf("call count = %d, want %d", len(rt.calls), MaxAttempts)
	}
	if len(spy.waits) != MaxAttempts-1 {
		t.Fatalf("sleep count = %d, want %d", len(spy.waits), MaxAttempts-1)
	}
}

func TestWrite_clampsRetryAfter_whenHeaderValueExceedsMax(t *testing.T) {
	t.Parallel()

	// Given: 過大な Retry-After 付き 429、その後成功応答
	w, _, spy := newFakeTextWriterWithSleepSpy(
		fakeClientResponse{
			status: http.StatusTooManyRequests,
			header: http.Header{"Retry-After": {"999999"}},
			body:   `{"error":"rate limited"}`,
		},
		successResponse("STOP", "クランプ後の原稿。"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 成功断片が返り、待ちは MaxRetryAfter でクランプされる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "クランプ後の原稿。" {
		t.Fatalf("Write() = %q", got)
	}
	if len(spy.waits) != 1 {
		t.Fatalf("sleep count = %d, want 1", len(spy.waits))
	}
	if spy.waits[0] != MaxRetryAfter {
		t.Fatalf("wait = %v, want %v (clamped)", spy.waits[0], MaxRetryAfter)
	}
}

func TestWrite_fallsBackToBackoff_when429RetryAfterIsUnparseable(t *testing.T) {
	t.Parallel()

	// Given: 解釈できない Retry-After 付き 429、その後成功応答
	w, _, spy := newFakeTextWriterWithSleepSpy(
		fakeClientResponse{
			status: http.StatusTooManyRequests,
			header: http.Header{"Retry-After": {"soon"}},
			body:   `{"error":"rate limited"}`,
		},
		successResponse("STOP", "backoff 復帰後の原稿。"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 成功断片が返り、待ちは backoffDelay(1) == 1s
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "backoff 復帰後の原稿。" {
		t.Fatalf("Write() = %q", got)
	}
	if len(spy.waits) != 1 || spy.waits[0] != time.Second {
		t.Fatalf("waits = %v, want [1s]", spy.waits)
	}
}

func TestWrite_doesNotRetry_whenClientErrorStatus(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			t.Parallel()

			// Given: client error status を返す
			w, rt := newFakeTextWriter(fakeClientResponse{status: status, body: `{"error":"denied"}`})

			// When: Write する
			got, err := w.Write(context.Background(), "原稿を書いて")

			// Then: 非 retry（calls == 1）で *adaptererror.Error、断片空
			assertGeminiInfraErrorOp(t, err, "http_status")
			if got != "" {
				t.Fatalf("fragment = %q, want empty", got)
			}
			if len(rt.calls) != 1 {
				t.Fatalf("call count = %d, want 1", len(rt.calls))
			}
			// geminiapi は secondary。番兵 port.ErrSourceExhausted を出す相手がいない。
			if errors.Is(err, port.ErrSourceExhausted) {
				t.Fatalf("errors.Is(err, port.ErrSourceExhausted) が true: %v", err)
			}
		})
	}
}

func TestWrite_includesResponseBodySnippet_whenStatus401(t *testing.T) {
	t.Parallel()

	// Given: 401 応答の body に切り分け用の理由（status / message）が入っている
	const body = `{"error":{"code":401,"status":"UNAUTHENTICATED","message":"API key not valid"}}`
	w, _ := newFakeTextWriter(fakeClientResponse{status: http.StatusUnauthorized, body: body})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて")

	// Then: http_status Infra Error に応答 body の snippet が載る（System 失敗の切り分け用）
	assertGeminiInfraErrorOp(t, err, "http_status")
	if !strings.Contains(err.Error(), "response body:") {
		t.Fatalf("error message %q does not carry response body snippet", err.Error())
	}
	if !strings.Contains(err.Error(), "UNAUTHENTICATED") {
		t.Fatalf("error message %q does not carry response body reason %q", err.Error(), "UNAUTHENTICATED")
	}
}

func TestWrite_returnsInfraError_whenFinishReasonIsNotStop(t *testing.T) {
	t.Parallel()

	// Given: finishReason が MAX_TOKENS の成功 status 応答
	w, rt := newFakeTextWriter(successResponse("MAX_TOKENS", "途中で切れた原稿。"))

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 非 retry で *adaptererror.Error、断片空
	assertGeminiInfraErrorOp(t, err, "finish_reason")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenTextIsWhitespaceOnly(t *testing.T) {
	t.Parallel()

	// Given: finishReason STOP だが text が空白のみ
	w, rt := newFakeTextWriter(successResponse("STOP", "   "))

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 非 retry で *adaptererror.Error、断片空
	assertGeminiInfraErrorOp(t, err, "empty_text")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenCandidatesEmpty(t *testing.T) {
	t.Parallel()

	// Given: candidates が空
	w, _ := newFakeTextWriter(fakeClientResponse{
		status: http.StatusOK,
		header: http.Header{"Content-Type": {"application/json"}},
		body:   `{"candidates":[]}`,
	})

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 非 retry で parse_response の *adaptererror.Error、断片空
	assertGeminiInfraErrorOp(t, err, "parse_response")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
}

func TestWrite_returnsInfraError_whenResponseBodyNotJSON(t *testing.T) {
	t.Parallel()

	// Given: 200 だが JSON でない body
	w, _ := newFakeTextWriter(fakeClientResponse{
		status: http.StatusOK,
		header: http.Header{"Content-Type": {"application/json"}},
		body:   `not json`,
	})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて")

	// Then: 非 retry の parse_response Infra Error
	assertGeminiInfraErrorOp(t, err, "parse_response")
}

func TestWrite_returnsInfraError_whenBriefEmptyAfterTrim(t *testing.T) {
	t.Parallel()

	// Given: trim 後空の brief。Client は呼ばれない想定
	w, rt := newFakeTextWriter()

	// When: Write する
	got, err := w.Write(context.Background(), "  \t\n  ")

	// Then: validate_brief Infra Error、断片空、Client は呼ばれない
	assertGeminiInfraErrorOp(t, err, "validate_brief")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
	if len(rt.calls) != 0 {
		t.Fatalf("call count = %d, want 0", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenClientIsNil(t *testing.T) {
	t.Parallel()

	// Given: client nil の TextWriter
	w := newTextWriter(nil, fakeAPIKey, func(context.Context, time.Duration) {})

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて")

	// Then: build_request Infra Error、断片空
	assertGeminiInfraErrorOp(t, err, "build_request")
	if got != "" {
		t.Fatalf("fragment = %q, want empty", got)
	}
}

func TestWrite_excludesAPIKeyFromErrorMessage_andPutsItInHeader(t *testing.T) {
	t.Parallel()

	// Given: 常に 400 を返す応答（fake key の実値は error に出てはならない）
	w, _ := newFakeTextWriter(fakeClientResponse{status: http.StatusBadRequest, body: `{"error":"INVALID_ARGUMENT"}`})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて")

	// Then: error は返るが apiKey 実値は message に出ない
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), fakeAPIKey) {
		t.Fatalf("error message %q contains api key value", err.Error())
	}

	// Given: 成功応答で request header を観測する
	w2, rt2 := newFakeTextWriter(successResponse("STOP", "header 検査用の原稿。"))

	// When: Write する
	if _, err := w2.Write(context.Background(), "原稿を書いて"); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: key は APIKeyHeader に載り、URL query には出ない
	if got := rt2.calls[0].Header.Get(APIKeyHeader); got != fakeAPIKey {
		t.Fatalf("%s = %q, want %q", APIKeyHeader, got, fakeAPIKey)
	}
	if strings.Contains(rt2.calls[0].URL, fakeAPIKey) {
		t.Fatalf("URL %q contains api key", rt2.calls[0].URL)
	}
	if strings.Contains(rt2.calls[0].URL, "key=") {
		t.Fatalf("URL %q carries a key query param", rt2.calls[0].URL)
	}
}

func TestNewTextWriter_returnsNonNil(t *testing.T) {
	t.Parallel()

	w := NewTextWriter(&http.Client{}, "gemini-fake-key")
	if w == nil {
		t.Fatal("NewTextWriter が nil を返した")
	}

	// 内部 constructor の backoff seam 分岐（nil → ctxSleep）も 1 度通す。
	if got := newTextWriter(&http.Client{}, "k", nil); got.backoffSleepFn == nil {
		t.Fatal("newTextWriter(nil) が backoffSleepFn を補わなかった")
	}
}

func TestCtxSleep_returnsWhenContextDone(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctxSleep(ctx, time.Hour) // ctx が済んでいるので即戻る
}
