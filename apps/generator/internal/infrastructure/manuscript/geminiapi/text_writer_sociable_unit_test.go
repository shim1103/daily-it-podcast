package geminiapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
)

// Scope: Sociable Unit
// 実物: geminiapi.TextWriter（Gemini generateContent Adapter）
// Double: http.RoundTripper の Spy（fakeRoundTripper）。backoffSleepFn は待ちを観測する sleepSpy。
// buildFn は raw response を渡す Stub（validBuildFn）、または渡された raw を記録する Spy
// （recordingBuildFn / rejectingBuildFn / sequenceBuildFn）。
//
// retry / error 方針: 成功応答 + buildFn 成功でその draft を返す / buildFn が invalid を返したら
// port.BuildRejectionBrief で組んだ brief で TextWriterMaxAttempts 回まで再試行 / 使い切ったら
// port.ErrDraftRejected を wrap し port.LastAttempt を chain へ含める / generateContent が retry
// しない error で抜けるとき、直前 attempt の raw/buildErr があれば port.LastAttempt として chain
// へ含め、無ければ透過する / HTTP 層の retry 方針（429 backoff・5xx 1 回・その他非 retry）は既存のまま。

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

// retryReporterCall は retryReporterSpy が観測した Retry 呼び出し 1 件分。
type retryReporterCall struct {
	step    string
	attempt int
	max     int
	reason  string
}

// retryReporterSpy は port.RetryReporter を満たし、Retry 呼び出しを記録する Spy。
type retryReporterSpy struct {
	calls []retryReporterCall
}

func (s *retryReporterSpy) Retry(step string, attempt, max int, reason string) {
	s.calls = append(s.calls, retryReporterCall{step: step, attempt: attempt, max: max, reason: reason})
}

func newFakeTextWriter(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper) {
	w, rt, _ := newFakeTextWriterWithSleepSpy(responses...)
	return w, rt
}

func newFakeTextWriterWithSleepSpy(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper, *sleepSpy) {
	w, rt, spy, _ := newFakeTextWriterWithSpies(responses...)
	return w, rt, spy
}

func newFakeTextWriterWithSpies(responses ...fakeClientResponse) (*TextWriter, *fakeRoundTripper, *sleepSpy, *retryReporterSpy) {
	rt := &fakeRoundTripper{responses: responses}
	spy := &sleepSpy{}
	retry := &retryReporterSpy{}
	w := newTextWriter(&http.Client{Transport: rt}, fakeAPIKey, TierFree, func(_ context.Context, d time.Duration) {
		spy.waits = append(spy.waits, d)
	}, retry)
	return w, rt, spy, retry
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

// validDraft は buildFn Stub / Spy が成功時に返す固定 draft。
var validDraft = models.ManuscriptDraft{Title: "valid draft"}

// validBuildFn は raw を検証せず常に validDraft を返す buildFn の Stub。
func validBuildFn(string) (models.ManuscriptDraft, error) {
	return validDraft, nil
}

// recordingBuildFn は受け取った raw を記録しつつ validBuildFn と同じ結果を返す buildFn の Spy。
type recordingBuildFn struct {
	raws []string
}

func (r *recordingBuildFn) call(raw string) (models.ManuscriptDraft, error) {
	r.raws = append(r.raws, raw)
	return validDraft, nil
}

// rejectingBuildFn は常に invalid 判定の error を返しつつ、呼ばれた raw を記録する buildFn の Spy。
type rejectingBuildFn struct {
	err  error
	raws []string
}

func (r *rejectingBuildFn) call(raw string) (models.ManuscriptDraft, error) {
	r.raws = append(r.raws, raw)
	return models.ManuscriptDraft{}, r.err
}

// sequenceBuildFn は呼び出し順に rejects 回だけ error を返し、以降は validDraft を返しつつ、
// 呼ばれた raw を記録する buildFn の Spy。
type sequenceBuildFn struct {
	rejects int
	err     error
	calls   int
	raws    []string
}

func (s *sequenceBuildFn) call(raw string) (models.ManuscriptDraft, error) {
	s.raws = append(s.raws, raw)
	s.calls++
	if s.calls <= s.rejects {
		return models.ManuscriptDraft{}, s.err
	}
	return validDraft, nil
}

// assertLastAttempt は err の chain から port.LastAttempt を取り出し、Raw/BuildErr が一致することを固定する。
func assertLastAttempt(t *testing.T, err error, wantRaw string, wantBuildErr error) {
	t.Helper()
	var la port.LastAttempt
	if !errors.As(err, &la) {
		t.Fatalf("errors.As(err, &port.LastAttempt{}) failed: %v", err)
	}
	if la.Raw != wantRaw {
		t.Fatalf("LastAttempt.Raw = %q, want %q", la.Raw, wantRaw)
	}
	if !errors.Is(la.BuildErr, wantBuildErr) {
		t.Fatalf("LastAttempt.BuildErr = %v, want %v", la.BuildErr, wantBuildErr)
	}
}

func TestWrite_returnsBuildFnDraft_whenGenerateContentAndBuildFnBothSucceedOnFirstAttempt(t *testing.T) {
	t.Parallel()

	// Given: 1 候補・finishReason STOP・非空 text を 2 part 返す成功応答（parts を連結する前提）。buildFn は 1 回目で成功する
	const brief = "本文の要約から原稿を書いて"
	w, rt := newFakeTextWriter(successResponse("STOP", "前半の原稿。", "後半の原稿。"))
	recorder := &recordingBuildFn{}

	// When: Write する
	got, err := w.Write(context.Background(), brief, recorder.call)

	// Then: buildFn が返した draft がそのまま返り、HTTP 呼び出しは 1 回、buildFn には連結済み raw が渡る
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
	if len(recorder.raws) != 1 || recorder.raws[0] != "前半の原稿。後半の原稿。" {
		t.Fatalf("buildFn raws = %q, want [%q]", recorder.raws, "前半の原稿。後半の原稿。")
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

func TestGenerateContent_sendsURLContextAndGoogleSearchTools_onEveryRequest(t *testing.T) {
	t.Parallel()

	// Given: 成功応答 1 件
	w, rt := newFakeTextWriter(successResponse("STOP", "tool 検証用の原稿。"))

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: request body の tools に url_context と google_search が空 object で両方入る
	// （Gemini API v1beta generateContent の実際の tools 形式。ai.google.dev/gemini-api/docs/
	// generate-content/url-context・generate-content/google-search の REST curl 例で確認済み）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
	wantTools := requestToolsOf(t, rt.calls[0].Body)
	if len(wantTools) != 2 {
		t.Fatalf("tools = %+v, want 2 entries", wantTools)
	}
	if _, ok := wantTools[0]["url_context"]; !ok {
		t.Fatalf("tools[0] = %+v, want url_context key", wantTools[0])
	}
	if len(wantTools[0]["url_context"]) != 0 {
		t.Fatalf("tools[0].url_context = %+v, want empty object", wantTools[0]["url_context"])
	}
	if _, ok := wantTools[1]["google_search"]; !ok {
		t.Fatalf("tools[1] = %+v, want google_search key", wantTools[1])
	}
	if len(wantTools[1]["google_search"]) != 0 {
		t.Fatalf("tools[1].google_search = %+v, want empty object", wantTools[1]["google_search"])
	}
}

// requestToolsOf は fakeRoundTripper が記録した request body から tools 配列を取り出す。
func requestToolsOf(t *testing.T, body []byte) []map[string]map[string]any {
	t.Helper()
	var reqBody struct {
		Tools []map[string]map[string]any `json:"tools"`
	}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	return reqBody.Tools
}

func TestWrite_retriesWithRejectionBrief_whenBuildFnReportsInvalidOnce(t *testing.T) {
	t.Parallel()

	// Given: 1 回目・2 回目とも成功応答。buildFn は 1 回目だけ invalid を返す
	const brief = "本文の要約から原稿を書いて"
	w, rt := newFakeTextWriter(
		successResponse("STOP", "1 回目の原稿。"),
		successResponse("STOP", "2 回目の原稿。"),
	)
	buildErr := errors.New("invalid draft")
	seq := &sequenceBuildFn{rejects: 1, err: buildErr}

	// When: Write する
	got, err := w.Write(context.Background(), brief, seq.call)

	// Then: 2 回目で成功。2 回目の HTTP request body には port.BuildRejectionBrief で組んだ brief が入る
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
	wantSecondBrief := port.BuildRejectionBrief(brief, "1 回目の原稿。", buildErr.Error())
	if got := requestBriefOf(t, rt.calls[1].Body); got != wantSecondBrief {
		t.Fatalf("2 回目 brief = %q, want %q", got, wantSecondBrief)
	}
}

func TestWrite_notifiesRetryReporter_whenBuildFnReportsInvalidOnce(t *testing.T) {
	t.Parallel()

	// Given: 1 回目・2 回目とも成功応答。buildFn は 1 回目だけ invalid を返す
	w, _, _, retry := newFakeTextWriterWithSpies(
		successResponse("STOP", "1 回目の原稿。"),
		successResponse("STOP", "2 回目の原稿。"),
	)
	buildErr := errors.New("invalid draft")
	seq := &sequenceBuildFn{rejects: 1, err: buildErr}

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", seq.call)

	// Then: 次 attempt が残っている 1 回目失敗時だけ Retry が 1 回通知される
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if len(retry.calls) != 1 {
		t.Fatalf("retry calls = %+v, want 1 call", retry.calls)
	}
	want := retryReporterCall{step: "write_manuscript_draft", attempt: 1, max: TextWriterMaxAttempts, reason: buildErr.Error()}
	if retry.calls[0] != want {
		t.Fatalf("retry.calls[0] = %+v, want %+v", retry.calls[0], want)
	}
}

func TestWrite_doesNotNotifyRetryReporter_onFinalAttempt_whenBuildFnFailsAllAttempts(t *testing.T) {
	t.Parallel()

	// Given: 全 attempt で成功応答だが buildFn は毎回 invalid を返す
	responses := make([]fakeClientResponse, 0, TextWriterMaxAttempts)
	for i := 0; i < TextWriterMaxAttempts; i++ {
		responses = append(responses, successResponse("STOP", fmt.Sprintf("%d 回目の原稿。", i+1)))
	}
	w, _, _, retry := newFakeTextWriterWithSpies(responses...)
	buildErr := errors.New("invalid draft")
	rejecting := &rejectingBuildFn{err: buildErr}

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", rejecting.call)

	// Then: TextWriterMaxAttempts - 1 回だけ Retry が通知される（最終 attempt は次が無いので呼ばない）
	if !errors.Is(err, port.ErrDraftRejected) {
		t.Fatalf("errors.Is(err, port.ErrDraftRejected) = false: %v", err)
	}
	if len(retry.calls) != TextWriterMaxAttempts-1 {
		t.Fatalf("retry calls = %+v, want %d calls", retry.calls, TextWriterMaxAttempts-1)
	}
	for i, call := range retry.calls {
		wantAttempt := i + 1
		if call.attempt != wantAttempt || call.max != TextWriterMaxAttempts || call.step != "write_manuscript_draft" {
			t.Fatalf("retry.calls[%d] = %+v, want attempt=%d max=%d step=write_manuscript_draft", i, call, wantAttempt, TextWriterMaxAttempts)
		}
	}
}

func TestGenerateContent_notifiesRetryReporter_on429BeforeEachBackoff(t *testing.T) {
	t.Parallel()

	// Given: 429 を MaxAttempts 回返し続ける
	responses := make([]fakeClientResponse, 0, MaxAttempts)
	for i := 0; i < MaxAttempts; i++ {
		responses = append(responses, fakeClientResponse{
			status: http.StatusTooManyRequests,
			body:   `{"error":"rate limited"}`,
		})
	}
	w, _, _, retry := newFakeTextWriterWithSpies(responses...)

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: backoff 前の MaxAttempts-1 回だけ Retry が通知される（使い切りの最終 1 回は通知しない）
	assertGeminiInfraErrorOp(t, err, "http_status")
	if len(retry.calls) != MaxAttempts-1 {
		t.Fatalf("retry calls = %+v, want %d calls", retry.calls, MaxAttempts-1)
	}
	for i, call := range retry.calls {
		wantAttempt := i + 1
		if call.attempt != wantAttempt || call.max != MaxAttempts || call.step != "generate_content" {
			t.Fatalf("retry.calls[%d] = %+v, want attempt=%d max=%d step=generate_content", i, call, wantAttempt, MaxAttempts)
		}
	}
}

func TestWrite_returnsDraftRejectedWithLastAttempt_whenBuildFnFailsAllAttempts(t *testing.T) {
	t.Parallel()

	// Given: 全 attempt で成功応答だが buildFn は毎回 invalid を返す
	responses := make([]fakeClientResponse, 0, TextWriterMaxAttempts)
	for i := 0; i < TextWriterMaxAttempts; i++ {
		responses = append(responses, successResponse("STOP", fmt.Sprintf("%d 回目の原稿。", i+1)))
	}
	w, rt := newFakeTextWriter(responses...)
	buildErr := errors.New("invalid draft")
	rejecting := &rejectingBuildFn{err: buildErr}

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", rejecting.call)

	// Then: TextWriterMaxAttempts 回すべて呼ばれ、port.ErrDraftRejected を errors.Is で判定でき、
	// 最後の attempt の raw/buildErr が port.LastAttempt として chain から取り出せる
	if !errors.Is(err, port.ErrDraftRejected) {
		t.Fatalf("errors.Is(err, port.ErrDraftRejected) = false: %v", err)
	}
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != TextWriterMaxAttempts {
		t.Fatalf("call count = %d, want %d", len(rt.calls), TextWriterMaxAttempts)
	}
	if len(rejecting.raws) != TextWriterMaxAttempts {
		t.Fatalf("buildFn calls = %d, want %d", len(rejecting.raws), TextWriterMaxAttempts)
	}
	wantLastRaw := fmt.Sprintf("%d 回目の原稿。", TextWriterMaxAttempts)
	assertLastAttempt(t, err, wantLastRaw, buildErr)
}

func TestWrite_wrapsLastAttempt_whenGenerateContentFailsNonRetryableAfterPriorInvalidBuild(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は成功応答で buildFn が invalid、2 回目は非 retry の 4xx
	w, rt := newFakeTextWriter(
		successResponse("STOP", "1 回目の原稿。"),
		fakeClientResponse{status: http.StatusBadRequest, body: `{"error":"INVALID_ARGUMENT"}`},
	)
	buildErr := errors.New("invalid draft")
	rejecting := &rejectingBuildFn{err: buildErr}

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", rejecting.call)

	// Then: generateContent の非 retry error が返り、直前 attempt の raw/buildErr が
	// port.LastAttempt として chain から取り出せる
	assertGeminiInfraErrorOp(t, err, "http_status")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
	assertLastAttempt(t, err, "1 回目の原稿。", buildErr)
}

func TestWrite_propagatesErrorWithoutLastAttempt_whenGenerateContentFailsNonRetryableOnFirstAttempt(t *testing.T) {
	t.Parallel()

	// Given: 1 回目から非 retry の 4xx（直前 attempt が存在しない）
	w, rt := newFakeTextWriter(fakeClientResponse{status: http.StatusBadRequest, body: `{"error":"INVALID_ARGUMENT"}`})

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: error はそのまま透過し、port.LastAttempt は chain に含まれない
	assertGeminiInfraErrorOp(t, err, "http_status")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
	var la port.LastAttempt
	if errors.As(err, &la) {
		t.Fatalf("errors.As(err, &port.LastAttempt{}) = true, want false: %v", err)
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: +1 即再試行で 2 回目成功（calls == 2）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: +1 即再試行で 2 回目成功（calls == 2）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 即再試行は 1 回だけ（calls == 2）で *adaptererror.Error
	assertGeminiInfraErrorOp(t, err, "http_status")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: read_body 断は一過性として 1 回だけ再試行し、2 回目で成功する（calls == 2）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 即再試行は 1 回だけ（calls == 2）で read_body の Infrastructure Error
	assertGeminiInfraErrorOp(t, err, "read_body")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
	}
}

func TestWrite_wrapsSourceExhausted_afterTierFreeUsesAll429Attempts(t *testing.T) {
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 上限到達で次 source へ切替可能な番兵と *adaptererror.Error、待ち観測は MaxAttempts-1 回、call は MaxAttempts 回
	assertGeminiInfraErrorOp(t, err, "http_status")
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v", err)
	}
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != MaxAttempts {
		t.Fatalf("call count = %d, want %d", len(rt.calls), MaxAttempts)
	}
	if len(spy.waits) != MaxAttempts-1 {
		t.Fatalf("sleep count = %d, want %d", len(spy.waits), MaxAttempts-1)
	}
}

func TestWrite_wrapsSourceExhausted_afterTierPaidUsesAll429Attempts(t *testing.T) {
	t.Parallel()

	// Given: paid source が 429 を MaxAttempts 回返し続ける
	responses := make([]fakeClientResponse, 0, MaxAttempts)
	for i := 0; i < MaxAttempts; i++ {
		responses = append(responses, fakeClientResponse{
			status: http.StatusTooManyRequests,
			body:   `{"error":"rate limited"}`,
		})
	}
	w, _, _ := newFakeTextWriterWithSleepSpy(responses...)
	w.tier = TierPaid

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: source の位置に依存せず切替用番兵と Infrastructure Error を返す
	assertGeminiInfraErrorOp(t, err, "http_status")
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) = false: %v", err)
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 成功 draft が返り、待ちは MaxRetryAfter でクランプされる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 成功 draft が返り、待ちは backoffDelay(1) == 1s
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, validDraft)
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
			got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

			// Then: 非 retry（calls == 1）で *adaptererror.Error、draft は zero value
			assertGeminiInfraErrorOp(t, err, "http_status")
			assertDraftEqual(t, got, models.ManuscriptDraft{})
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
	_, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 非 retry で *adaptererror.Error、draft は zero value
	assertGeminiInfraErrorOp(t, err, "finish_reason")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenTextIsWhitespaceOnly(t *testing.T) {
	t.Parallel()

	// Given: finishReason STOP だが text が空白のみ
	w, rt := newFakeTextWriter(successResponse("STOP", "   "))

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 非 retry で *adaptererror.Error、draft は zero value
	assertGeminiInfraErrorOp(t, err, "empty_text")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
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
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 非 retry で parse_response の *adaptererror.Error、draft は zero value
	assertGeminiInfraErrorOp(t, err, "parse_response")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
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
	_, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: 非 retry の parse_response Infra Error
	assertGeminiInfraErrorOp(t, err, "parse_response")
}

func TestWrite_returnsInfraError_whenBriefEmptyAfterTrim(t *testing.T) {
	t.Parallel()

	// Given: trim 後空の brief。Client は呼ばれない想定
	w, rt := newFakeTextWriter()

	// When: Write する
	got, err := w.Write(context.Background(), "  \t\n  ", validBuildFn)

	// Then: validate_brief Infra Error、draft は zero value、Client は呼ばれない
	assertGeminiInfraErrorOp(t, err, "validate_brief")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if len(rt.calls) != 0 {
		t.Fatalf("call count = %d, want 0", len(rt.calls))
	}
}

func TestWrite_returnsInfraError_whenClientIsNil(t *testing.T) {
	t.Parallel()

	// Given: client nil の TextWriter
	w := newTextWriter(nil, fakeAPIKey, TierFree, func(context.Context, time.Duration) {}, nil)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

	// Then: build_request Infra Error、draft は zero value
	assertGeminiInfraErrorOp(t, err, "build_request")
	assertDraftEqual(t, got, models.ManuscriptDraft{})
}

func TestWrite_excludesAPIKeyFromErrorMessage_andPutsItInHeader(t *testing.T) {
	t.Parallel()

	// Given: 常に 400 を返す応答（fake key の実値は error に出てはならない）
	w, _ := newFakeTextWriter(fakeClientResponse{status: http.StatusBadRequest, body: `{"error":"INVALID_ARGUMENT"}`})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", validBuildFn)

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
	if _, err := w2.Write(context.Background(), "原稿を書いて", validBuildFn); err != nil {
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

	w := NewTextWriter(&http.Client{}, "gemini-fake-key", TierFree, nil)
	if w == nil {
		t.Fatal("NewTextWriter が nil を返した")
	}

	// 内部 constructor の backoff seam 分岐（nil → ctxSleep）も 1 度通す。
	if got := newTextWriter(&http.Client{}, "k", TierFree, nil, nil); got.backoffSleepFn == nil {
		t.Fatal("newTextWriter(nil) が backoffSleepFn を補わなかった")
	}
}

// TestNewTextWriter_storesGivenTier_whenConstructed は渡した Tier がそのまま保持されることを
// 検証する（geminiapi.Tier は gemini（TTS）の Tier と同型の識別子。値の使用先は将来 task）。
func TestNewTextWriter_storesGivenTier_whenConstructed(t *testing.T) {
	t.Parallel()

	// Given / When: TierPaid を渡して構築する
	w := NewTextWriter(&http.Client{}, "gemini-fake-key", TierPaid, nil)

	// Then: 渡した Tier がそのまま保持される
	if w.tier != TierPaid {
		t.Fatalf("tier = %v, want %v", w.tier, TierPaid)
	}
}

func TestCtxSleep_returnsWhenContextDone(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctxSleep(ctx, time.Hour) // ctx が済んでいるので即戻る
}

// requestBriefOf は fakeRoundTripper が記録した request body から contents[0].parts[0].text を取り出す。
func requestBriefOf(t *testing.T, body []byte) string {
	t.Helper()
	var reqBody struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(body, &reqBody); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if len(reqBody.Contents) != 1 || len(reqBody.Contents[0].Parts) != 1 {
		t.Fatalf("request body shape = %+v", reqBody)
	}
	return reqBody.Contents[0].Parts[0].Text
}

// assertDraftEqual は got が want と同一 draft であることを固定する。
// models.ManuscriptDraft は slice field を持つため == 比較不可であり reflect.DeepEqual を使う。
func assertDraftEqual(t *testing.T, got, want models.ManuscriptDraft) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Write() = %+v, want %+v", got, want)
	}
}
