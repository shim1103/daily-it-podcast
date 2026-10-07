package cursorapi

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
// 実物: cursorapi.TextWriter（Cursor Cloud Agents Adapter）
// Double: http.RoundTripper の Spy（fakeRoundTripper）。backoffSleepFn は待ちを観測する sleepSpy。
// buildFn は Stub（valid/invalid を呼び出し回数で切り替える）。
//
// retry 方針: 1 回目は createAgent（POST /v1/agents）で agent と run を同時 create。buildFn が
// invalid を返したら 2 回目以降は同じ agentId へ createRun（POST /v1/agents/{id}/runs）で
// follow-up run を送り、新しい runId の stream から再取得する（TextWriterMaxAttempts 回まで）。
// stream 取得の 429/5xx/Do error retry は既存の streamResult 方針のまま。

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
	// why: このヘルパーは stream retry / buildFn reject retry の両方を発生させる test から
	//      共有される。呼ばれるかどうかをテストごとに見極めず、常に Spy を渡して安全に倒す。
	w := newTextWriter(&http.Client{Transport: rt}, "cursor-fake-key", func(_ context.Context, d time.Duration) {
		spy.waits = append(spy.waits, d)
	}, &retryReporterSpy{})
	return w, rt, spy
}

// retryReporterSpy は port.RetryReporter を満たし、Retry 呼び出しを記録する Spy。
type retryReporterSpy struct {
	calls int
}

func (s *retryReporterSpy) Retry(step string, attempt, max int, reason string) {
	s.calls++
}

// alwaysValidBuildFn は常に valid と判定する Stub。retry させたくない test で使う。
func alwaysValidBuildFn(raw string) (models.ManuscriptDraft, error) {
	return models.ManuscriptDraft{Title: raw}, nil
}

// rejectNTimesBuildFn は先頭 n 回を invalid（rejectErr）と判定し、n+1 回目以降を valid とする Stub。
// invalid-draft retry ループが buildFn の判定に従って正しく回数を切り替えることを検証する。
func rejectNTimesBuildFn(n int, rejectErr error) func(string) (models.ManuscriptDraft, error) {
	calls := 0
	return func(raw string) (models.ManuscriptDraft, error) {
		calls++
		if calls <= n {
			return models.ManuscriptDraft{}, rejectErr
		}
		return models.ManuscriptDraft{Title: raw}, nil
	}
}

// alwaysInvalidBuildFn は常に invalid と判定する Stub。retry 上限到達を検証する test で使う。
func alwaysInvalidBuildFn(rejectErr error) func(string) (models.ManuscriptDraft, error) {
	return func(string) (models.ManuscriptDraft, error) {
		return models.ManuscriptDraft{}, rejectErr
	}
}

// createAgentBody は create 応答の fixture を組む。
func createAgentBody(agentID, runID string) string {
	raw, _ := json.Marshal(map[string]any{
		"agent": map[string]any{"id": agentID, "status": "ACTIVE"},
		"run":   map[string]any{"id": runID, "agentId": agentID, "status": "CREATING"},
	})
	return string(raw)
}

// createRunBody は follow-up run create 応答の fixture を組む。
func createRunBody(runID string) string {
	raw, _ := json.Marshal(map[string]any{"id": runID, "status": "CREATING"})
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
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
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
	var infra *adaptererror.Error
	_ = errors.As(err, &infra)
	if infra.Op != wantOp {
		t.Fatalf("Op = %q, want %q", infra.Op, wantOp)
	}
}

func assertSourceExhausted(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) が false: %v", err)
	}
}

func assertZeroDraft(t *testing.T, got models.ManuscriptDraft) {
	t.Helper()
	if !reflect.DeepEqual(got, models.ManuscriptDraft{}) {
		t.Fatalf("draft = %+v, want zero value", got)
	}
}

func TestWrite_returnsDraft_whenCreateThenStreamSucceeds(t *testing.T) {

	// Given: create 成功と、終端 result に非空 text を持つ SSE を返す Client Stub。buildFn は常に valid
	const fragment = "本日の IT ニュース原稿の断片"
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		successStreamResponse(fragment),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "本文の要約から原稿を書いて", alwaysValidBuildFn)

	// Then: buildFn が返す draft が返り、create は POST、stream は GET で叩かれる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got.Title != fragment {
		t.Fatalf("Write().Title = %q, want %q", got.Title, fragment)
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
	got, err := w.Write(context.Background(), "  \t\n  ", alwaysValidBuildFn)

	// Then: validate_brief 系 Infra Error、draft ゼロ値、Client は呼ばれない
	assertCursorInfraErrorOp(t, err, "validate_brief")
	assertZeroDraft(t, got)
	if len(rt.calls) != 0 {
		t.Fatalf("call count = %d, want 0", len(rt.calls))
	}
}

func TestWrite_sendsFollowUpRun_whenBuildFnRejectsFirstAttempt(t *testing.T) {

	// Given: create 成功 → 1 回目 stream は invalid 判定される断片 → follow-up run create 成功
	//        → 2 回目 stream で valid な断片
	rejectErr := errors.New("topics count is 2, want 3")
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		successStreamResponse("1 回目の断片"),
		fakeClientResponse{status: http.StatusOK, body: createRunBody("run-2")},
		successStreamResponse("2 回目の断片"),
	)

	// When: Write する（buildFn は 1 回目だけ reject）
	got, err := w.Write(context.Background(), "原稿を書いて", rejectNTimesBuildFn(1, rejectErr))

	// Then: 2 回目の断片で valid になり、agentId は変えず createRun（POST .../v1/agents/bc-1/runs）を叩く
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got.Title != "2 回目の断片" {
		t.Fatalf("Write().Title = %q, want %q", got.Title, "2 回目の断片")
	}
	if len(rt.calls) != 4 {
		t.Fatalf("call count = %d, want 4 (create, stream, createRun, stream)", len(rt.calls))
	}
	if rt.calls[2].Method != http.MethodPost {
		t.Fatalf("createRun method = %q, want POST", rt.calls[2].Method)
	}
	if rt.calls[2].URL != APIBaseURL+AgentsPath+"/bc-1/runs" {
		t.Fatalf("createRun URL = %q, want %q", rt.calls[2].URL, APIBaseURL+AgentsPath+"/bc-1/runs")
	}
	if !strings.Contains(rt.calls[3].URL, "/v1/agents/bc-1/runs/run-2/stream") {
		t.Fatalf("2nd stream URL = %q, want .../v1/agents/bc-1/runs/run-2/stream", rt.calls[3].URL)
	}
	var runReqBody map[string]any
	if err := json.Unmarshal(rt.calls[2].Body, &runReqBody); err != nil {
		t.Fatalf("decode createRun request: %v", err)
	}
	prompt, _ := runReqBody["prompt"].(map[string]any)
	promptText, _ := prompt["text"].(string)
	// why: follow-up run は agent との会話継続なので、rejection 理由だけを新しい prompt として送る。
	if strings.Contains(promptText, "原稿を書いて") {
		t.Fatalf("createRun prompt.text = %q は元 brief を含んではならない", promptText)
	}
	if !strings.Contains(promptText, rejectErr.Error()) {
		t.Fatalf("createRun prompt.text = %q, want rejection reason %q を含む", promptText, rejectErr.Error())
	}
	if _, hasModel := runReqBody["model"]; hasModel {
		t.Fatalf("createRun request has model field, want follow-up run without model override")
	}
}

func TestWrite_wrapsDraftRejected_whenBuildFnRejectsAllAttempts(t *testing.T) {

	// Given: create 成功後、毎回 stream は成功するが buildFn は毎回 invalid と判定する
	rejectErr := errors.New("json is malformed")
	responses := []fakeClientResponse{
		{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		successStreamResponse("断片1"),
	}
	for i := 2; i <= TextWriterMaxAttempts; i++ {
		responses = append(responses,
			fakeClientResponse{status: http.StatusOK, body: createRunBody(fmt.Sprintf("run-%d", i))},
			successStreamResponse(fmt.Sprintf("断片%d", i)),
		)
	}
	w, rt := newFakeTextWriter(responses...)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysInvalidBuildFn(rejectErr))

	// Then: TextWriterMaxAttempts 回使い切って port.ErrDraftRejected で wrap、LastAttempt を辿れる
	if !errors.Is(err, port.ErrDraftRejected) {
		t.Fatalf("errors.Is(err, port.ErrDraftRejected) が false: %v", err)
	}
	assertZeroDraft(t, got)
	var lastAttempt port.LastAttempt
	if !errors.As(err, &lastAttempt) {
		t.Fatalf("errors.As(err, &port.LastAttempt{}) が false: %v", err)
	}
	if lastAttempt.Raw != fmt.Sprintf("断片%d", TextWriterMaxAttempts) {
		t.Fatalf("LastAttempt.Raw = %q, want %q", lastAttempt.Raw, fmt.Sprintf("断片%d", TextWriterMaxAttempts))
	}
	if !errors.Is(lastAttempt.BuildErr, rejectErr) {
		t.Fatalf("LastAttempt.BuildErr = %v, want %v", lastAttempt.BuildErr, rejectErr)
	}
	// create 1 回 + (stream + createRun) を繰り返し、最後は createRun なしで stream のみ
	// = 1 (create) + TextWriterMaxAttempts (stream) + (TextWriterMaxAttempts-1) (createRun)
	wantCalls := 1 + TextWriterMaxAttempts + (TextWriterMaxAttempts - 1)
	if len(rt.calls) != wantCalls {
		t.Fatalf("call count = %d, want %d", len(rt.calls), wantCalls)
	}
}

func TestWrite_returnsLastAttempt_whenCreateRunFailsAfterFirstRejection(t *testing.T) {

	// Given: create 成功 → 1 回目 stream は invalid 判定される断片 → createRun が 5xx を返し続ける
	rejectErr := errors.New("topics count is 2, want 3")
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		successStreamResponse("1 回目の断片"),
		fakeClientResponse{status: http.StatusInternalServerError, body: `{"error":"internal"}`},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", rejectNTimesBuildFn(1, rejectErr))

	// Then: createRun は非 idempotent なので再試行せず即座に Infra Error。直前 attempt の
	//       LastAttempt（断片・rejectErr）を chain へ持ち越す
	assertCursorInfraErrorOp(t, err, "create_run_status")
	assertZeroDraft(t, got)
	var lastAttempt port.LastAttempt
	if !errors.As(err, &lastAttempt) {
		t.Fatalf("errors.As(err, &port.LastAttempt{}) が false: %v", err)
	}
	if lastAttempt.Raw != "1 回目の断片" {
		t.Fatalf("LastAttempt.Raw = %q, want %q", lastAttempt.Raw, "1 回目の断片")
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(rt.calls))
	}
}

func TestWrite_retriesStreamOn429WithRetryAfter_untilMaxAttemptsThenSourceExhausted(t *testing.T) {

	// Given: create 成功後、stream 取得が Retry-After 付きの 429 を返し続ける
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 上限到達で次 source へ渡せる番兵と Infra Error、stream 取得は MaxAttempts 回
	assertCursorInfraErrorOp(t, err, "stream_status")
	assertSourceExhausted(t, err)
	assertZeroDraft(t, got)
	streamCalls := len(rt.calls) - 1
	if streamCalls != MaxAttempts {
		t.Fatalf("stream call count = %d, want %d", streamCalls, MaxAttempts)
	}
}

func TestWrite_retriesStreamOn429WithRetryAfter_thenSucceeds(t *testing.T) {

	// Given: create 成功後、stream 取得が Retry-After 付きの 429 を 1 回、2 回目で成功 SSE
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusTooManyRequests, header: http.Header{"Retry-After": {"1"}}, body: `{"error":"rate limited"}`},
		successStreamResponse("再試行後の断片"),
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 2 回目の stream 取得で非空断片
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got.Title != "再試行後の断片" {
		t.Fatalf("Write().Title = %q, want %q", got.Title, "再試行後の断片")
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(rt.calls))
	}
}

func TestWrite_wrapsSourceExhaustedWithoutWaiting_whenStream429DeclaresNoRecovery(t *testing.T) {

	// Given: create 成功後、stream 取得が回復の明示（解釈できる Retry-After）の無い 429 を返す。body は読まない
	w, rt, spy := newFakeTextWriterWithSleepSpy(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusTooManyRequests, body: `{"error":"quota"}`},
	)

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: retry も待ちもせず次 source へ渡せる番兵を返す（stream 取得は 1 回）
	assertCursorInfraErrorOp(t, err, "stream_status")
	assertSourceExhausted(t, err)
	if streamCalls := len(rt.calls) - 1; streamCalls != 1 {
		t.Fatalf("stream call count = %d, want 1", streamCalls)
	}
	if len(spy.waits) != 0 {
		t.Fatalf("waits = %v, want none", spy.waits)
	}
}

func TestWrite_wrapsSourceExhausted_whenCreateStatusIs429(t *testing.T) {

	// Given: create が 429 を返す（POST は非 retry。body は読まない）
	w, rt := newFakeTextWriter(fakeClientResponse{status: http.StatusTooManyRequests, body: `{"error":"x"}`})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 次 source へ渡せる番兵で wrap し、create は 1 回だけ
	assertSourceExhausted(t, err)
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestWrite_wrapsSourceExhaustedAfterOneRetry_whenStreamDoErrorPersists(t *testing.T) {

	// Given: create 成功後、stream 取得が Do error を 2 回返す（client 起因か server 起因かは区別できない）
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{err: fmt.Errorf("connection reset")},
		fakeClientResponse{err: fmt.Errorf("connection reset")},
	)

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 即再試行は 1 回だけ（stream 取得は 2 回）で、次 source へ渡せる番兵と do の Infra Error
	assertCursorInfraErrorOp(t, err, "do")
	assertSourceExhausted(t, err)
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3 (create + stream x2)", len(rt.calls))
	}
}

func TestWrite_wrapsSourceExhaustedWithoutRetry_whenCreateDoError(t *testing.T) {

	// Given: create（POST）が Do error を返す
	w, rt := newFakeTextWriter(fakeClientResponse{err: fmt.Errorf("connection reset")})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 非 idempotent なので再試行せず（呼び出し 1 回）、次 source へ渡せる番兵と do の Infra Error
	assertCursorInfraErrorOp(t, err, "do")
	assertSourceExhausted(t, err)
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: +1 即再試行で 2 回目成功、呼び出しは create + stream 2 回
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got.Title != "Do error 復帰後の断片" {
		t.Fatalf("Write().Title = %q", got.Title)
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: +1 即再試行で 2 回目成功
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got.Title != "5xx 復帰後の断片" {
		t.Fatalf("Write().Title = %q", got.Title)
	}
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(rt.calls))
	}
}

func TestWrite_wrapsSourceExhaustedAfterOneRetry_whenStream5xxPersists(t *testing.T) {

	// Given: create 成功後、stream 取得が 5xx を返し続ける
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusBadGateway, body: `{"error":"bad gateway"}`},
		fakeClientResponse{status: http.StatusBadGateway, body: `{"error":"bad gateway"}`},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 即再試行は 1 回だけ（stream 取得は 2 回）で、次 source へ渡せる番兵と Infra Error
	assertCursorInfraError(t, err)
	assertSourceExhausted(t, err)
	assertZeroDraft(t, got)
	if len(rt.calls) != 3 {
		t.Fatalf("call count = %d, want 3 (create + stream x2)", len(rt.calls))
	}
}

func TestWrite_wrapsSourceExhaustedWithoutRetry_whenCreateStatus5xx(t *testing.T) {

	// Given: create（POST /v1/agents）が 5xx を返す
	w, rt := newFakeTextWriter(fakeClientResponse{
		status: http.StatusInternalServerError,
		body:   `{"error":"internal"}`,
	})

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 非 idempotent なので再試行せず（呼び出し 1 回）、次 source へ渡せる番兵と Infra Error
	assertCursorInfraErrorOp(t, err, "create_status")
	assertSourceExhausted(t, err)
	assertZeroDraft(t, got)
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestClassifyStreamStatus_returnsRetryKindAndWait_perStatusAndRetryAfter(t *testing.T) {

	// Given: stream 取得の失敗 status と Retry-After の組
	cases := []struct {
		name     string
		status   int
		header   http.Header
		wantKind streamRetryKind
		wantWait time.Duration
	}{
		{"429 Retry-After 付き", http.StatusTooManyRequests, http.Header{"Retry-After": {"3"}}, retryRateLimited, 3 * time.Second},
		{"429 Retry-After 過大はクランプ", http.StatusTooManyRequests, http.Header{"Retry-After": {"999999"}}, retryRateLimited, MaxRetryAfter},
		{"429 Retry-After 無し", http.StatusTooManyRequests, nil, retryFallback, 0},
		{"429 Retry-After 解釈不能", http.StatusTooManyRequests, http.Header{"Retry-After": {"soon"}}, retryFallback, 0},
		{"500", http.StatusInternalServerError, nil, retryOnce, 0},
		{"502", http.StatusBadGateway, nil, retryOnce, 0},
		{"400 は bug", http.StatusBadRequest, nil, retryNone, 0},
		{"401 は credential 失効", http.StatusUnauthorized, nil, retryFallback, 0},
		{"403 は credential 失効", http.StatusForbidden, nil, retryFallback, 0},
		{"404", http.StatusNotFound, nil, retryNone, 0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			// When: 分類する
			got := classifyStreamStatus(tc.status, tc.header)

			// Then: 方針と待ちが決まる
			if got.kind != tc.wantKind || got.wait != tc.wantWait {
				t.Fatalf("classifyStreamStatus(%d, %v) = (%v, %v), want (%v, %v)", tc.status, tc.header, got.kind, got.wait, tc.wantKind, tc.wantWait)
			}
		})
	}
}

func TestShouldFallbackOnCreateStatus_returnsTrue_forCredentialRateLimitServerAndUsageLimit(t *testing.T) {

	// Given: create（非 idempotent）の失敗 status と body の組
	cases := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{"401 は credential 失効", http.StatusUnauthorized, `{}`, true},
		{"403 は credential 失効", http.StatusForbidden, `{}`, true},
		{"429 は待たずに fallback へ渡す", http.StatusTooManyRequests, `{}`, true},
		{"500 は fallback へ渡す", http.StatusInternalServerError, `{}`, true},
		{"503 は fallback へ渡す", http.StatusServiceUnavailable, `{}`, true},
		{"400 + usage_limit_exceeded は利用枠喪失", http.StatusBadRequest, `{"error":{"code":"usage_limit_exceeded"}}`, true},
		{"400 は bug", http.StatusBadRequest, `{"error":"malformed"}`, false},
		{"404 は bug", http.StatusNotFound, `{}`, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			// When: 判定する
			got := shouldFallbackOnCreateStatus(tc.status, []byte(tc.body))

			// Then: fallback へ渡すかどうかが決まる
			if got != tc.want {
				t.Fatalf("shouldFallbackOnCreateStatus(%d, %s) = %v, want %v", tc.status, tc.body, got, tc.want)
			}
		})
	}
}

func TestWrite_includesResponseBodySnippet_whenCreateStatusNotOK(t *testing.T) {

	// Given: create が 400 を返し、body に切り分け用の理由が入っている
	const reason = "your plan does not include background agents"
	w, _ := newFakeTextWriter(fakeClientResponse{
		status: http.StatusBadRequest,
		body:   `{"error":{"message":"` + reason + `"}}`,
	})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: create_status Infra Error に応答 body の snippet が載る（System 失敗の切り分け用）
	assertCursorInfraErrorOp(t, err, "create_status")
	if !strings.Contains(err.Error(), reason) {
		t.Fatalf("error message %q does not carry response body reason %q", err.Error(), reason)
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 非 retry の Infra Error（stream 取得は 1 回だけ）
	assertCursorInfraErrorOp(t, err, "empty_text")
	assertZeroDraft(t, got)
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: run 終端 error は非 retry の Infra Error
	assertCursorInfraErrorOp(t, err, "run_status")
	assertZeroDraft(t, got)
}

func TestWrite_excludesAPIKeyFromErrorMessage_whenCreateFails(t *testing.T) {

	// Given: create が 401 を返す。fake key の実値は error に出てはならない
	const apiKey = "cursor-secret-must-not-leak"
	rt := &fakeRoundTripper{responses: []fakeClientResponse{
		{status: http.StatusUnauthorized, body: `{"error":"unauthorized"}`},
	}}
	w := newTextWriter(&http.Client{Transport: rt}, apiKey, func(context.Context, time.Duration) {}, nil)

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

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
			got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

			// Then: 非 retry の Infra Error、draft ゼロ値、呼び出しは 1 回だけ
			assertCursorInfraErrorOp(t, err, "create_status")
			assertZeroDraft(t, got)
			if len(rt.calls) != 1 {
				t.Fatalf("call count = %d, want 1", len(rt.calls))
			}
		})
	}
}

func TestWrite_wrapsSourceExhausted_whenCreateStatusIs401Or403(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {

			// Given: create が 401 / 403（API key / subscription 失効）を返す
			w, _ := newFakeTextWriter(fakeClientResponse{status: status, body: `{"error":"denied"}`})

			// When: Write する
			_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

			// Then: vendor 非依存の番兵で wrap され、中身は *adaptererror.Error のまま辿れる
			assertSourceExhausted(t, err)
			assertCursorInfraErrorOp(t, err, "create_status")
		})
	}
}

func TestWrite_wrapsSourceExhausted_whenCreateStatusIs400WithUsageLimitExceeded(t *testing.T) {

	// Given: create が 400 かつ body に usage_limit_exceeded（Background Agent 利用枠喪失。run 34132953055 で実証）
	const body = `{"error":{"code":"usage_limit_exceeded","message":"Usage-based pricing required. Background Agent requires at least $2 remaining until your hard limit."}}`
	w, _ := newFakeTextWriter(fakeClientResponse{status: http.StatusBadRequest, body: body})

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 401/403 と同じく番兵で wrap され、中身は *adaptererror.Error のまま辿れる
	assertSourceExhausted(t, err)
	assertCursorInfraErrorOp(t, err, "create_status")
}

func TestWrite_wrapsSourceExhausted_whenCreateRunStatusIs401Or403(t *testing.T) {

	// Given: create 成功 → 1 回目は invalid 判定 → follow-up run（createRun）が 401 を返す
	rejectErr := errors.New("topics count is 2, want 3")
	w, _ := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		successStreamResponse("1 回目の断片"),
		fakeClientResponse{status: http.StatusUnauthorized, body: `{"error":"unauthorized"}`},
	)

	// When: Write する
	_, err := w.Write(context.Background(), "原稿を書いて", rejectNTimesBuildFn(1, rejectErr))

	// Then: createAgent と同じ判定基準（vendor 非依存の番兵）を createRun にも適用する
	assertSourceExhausted(t, err)
	assertCursorInfraErrorOp(t, err, "create_run_status")
}

func TestWrite_doesNotWrapSourceExhausted_whenCreateStatusIsBug(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusNotFound} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {

			// Given: create が 400 / 404 を返す（bug であり fallback へ渡さない。400 の body に usage_limit_exceeded は無い）
			w, _ := newFakeTextWriter(fakeClientResponse{status: status, body: `{"error":"x"}`})

			// When: Write する
			_, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

			// Then: 番兵で wrap しない
			if errors.Is(err, port.ErrSourceExhausted) {
				t.Fatalf("status %d をfallback へ渡した: %v", status, err)
			}
		})
	}
}

func TestWrite_wrapsSourceExhaustedWithoutRetry_whenStreamCredentialIsRejected(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		status := status
		t.Run(strconv.Itoa(status), func(t *testing.T) {

			// Given: create 成功後、stream 取得が API key の失効を示す status を返す
			w, rt := newFakeTextWriter(
				fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
				fakeClientResponse{status: status, body: `{"error":"denied"}`},
			)

			// When: Write する
			got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

			// Then: 非 retry で、次 source へ渡せる番兵と Infra Error。呼び出しは create + stream の 2 回だけ
			assertCursorInfraErrorOp(t, err, "stream_status")
			assertSourceExhausted(t, err)
			assertZeroDraft(t, got)
			if len(rt.calls) != 2 {
				t.Fatalf("call count = %d, want 2", len(rt.calls))
			}
		})
	}
}

func TestWrite_doesNotRetryOrWrapStream_whenRequestIsMalformed(t *testing.T) {

	// Given: create 成功後、stream 取得が呼び出し側の誤りを示す 400 を返す
	w, rt := newFakeTextWriter(
		fakeClientResponse{status: http.StatusOK, body: createAgentBody("bc-1", "run-1")},
		fakeClientResponse{status: http.StatusBadRequest, body: `{"error":"bad request"}`},
	)

	// When: Write する
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 非 retry で、bug として番兵で wrap せず返す。呼び出しは create + stream の 2 回だけ
	assertCursorInfraErrorOp(t, err, "stream_status")
	if errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("errors.Is(err, port.ErrSourceExhausted) が true: %v", err)
	}
	assertZeroDraft(t, got)
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(rt.calls))
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: parse_sse Infra Error。再 stream しない（呼び出しは create + stream の 2 回だけ）
	assertCursorInfraErrorOp(t, err, "parse_sse")
	assertZeroDraft(t, got)
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
	got, err := w.Write(context.Background(), "原稿を書いて", alwaysValidBuildFn)

	// Then: 成功断片が返り、待ち時間は MaxRetryAfter でクランプされる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got.Title != "クランプ後の断片" {
		t.Fatalf("Write().Title = %q", got.Title)
	}
	if len(spy.waits) != 1 {
		t.Fatalf("sleep count = %d, want 1", len(spy.waits))
	}
	if spy.waits[0] != MaxRetryAfter {
		t.Fatalf("wait = %v, want %v (clamped)", spy.waits[0], MaxRetryAfter)
	}
}
