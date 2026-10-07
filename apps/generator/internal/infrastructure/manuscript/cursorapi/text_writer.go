package cursorapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpdiag"
)

var _ port.TextWriter = (*TextWriter)(nil)

// TextWriter は Cursor Cloud Agents API を使う原稿 Adapter。
type TextWriter struct {
	client         *http.Client
	apiKey         string
	backoffSleepFn func(context.Context, time.Duration) // why: test の並列実行と共存するため package global に置かない
	retry          port.RetryReporter
}

// NewTextWriter は Cursor Cloud Agents API 用 TextWriter を組み立てる。
//
// @require apiKey は Composition で検証済み。retry != nil（Composition Root の結線責務）。
// @ensure 戻りは port.TextWriter。apiKey は Authorization: Bearer header にだけ使う。
// @ensure client == nil のとき Write は infraErr("build_request") を返す。
func NewTextWriter(client *http.Client, apiKey string, retry port.RetryReporter) *TextWriter {
	return newTextWriter(client, apiKey, ctxSleep, retry)
}

// why: NewTextWriter は本番の ctxSleep を固定し、test は待ちを観測する fake を渡す。
func newTextWriter(client *http.Client, apiKey string, backoffSleepFn func(context.Context, time.Duration), retry port.RetryReporter) *TextWriter {
	if backoffSleepFn == nil {
		backoffSleepFn = ctxSleep
	}
	return &TextWriter{client: client, apiKey: apiKey, backoffSleepFn: backoffSleepFn, retry: retry}
}

func ctxSleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Write は brief から原稿断片を得て、buildFn で ManuscriptDraft へ解釈する。
//
// @require brief は trim 後に非空。buildFn は非 nil。
// @ensure 成功時は buildFn が返す非 nil な models.ManuscriptDraft を返す。
// @ensure buildFn が invalid と判定した後の取得 error には port.LastAttempt を chain に含める。
// @ensure buildFn が TextWriterMaxAttempts 回とも error を返したら、port.ErrDraftRejected・最後の error・port.LastAttempt を wrap して返す。
// @ensure 取得元が当面使えないと分かった error は port.ErrSourceExhausted を wrap して返す（fallback へ渡す）。
// @invariant buildFn 呼び出しは 1 attempt につき高々 1 回。secret 実値は error へ出さない。
func (w *TextWriter) Write(ctx context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	if w == nil || w.client == nil {
		return models.ManuscriptDraft{}, infraErr("build_request", fmt.Errorf("client is nil"))
	}
	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return models.ManuscriptDraft{}, infraErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	var agentID string
	var last port.LastAttempt
	for attempt := 1; attempt <= TextWriterMaxAttempts; attempt++ {
		var raw string
		var err error
		agentID, raw, err = w.fetchFragment(ctx, agentID, trimmed, last)
		if err != nil {
			return models.ManuscriptDraft{}, withLastAttempt(err, last)
		}
		draft, err := buildFn(raw)
		if err == nil {
			return draft, nil
		}
		last = port.LastAttempt{Raw: raw, BuildErr: err}
		if attempt < TextWriterMaxAttempts {
			w.retry.Retry("write_manuscript_draft", attempt, TextWriterMaxAttempts, err.Error())
		}
	}
	return models.ManuscriptDraft{}, fmt.Errorf("%w: %w: %w", port.ErrDraftRejected, last.BuildErr, last)
}

// withLastAttempt は直前 attempt が invalid だった時だけ、その記録を err の chain へ足す。
func withLastAttempt(err error, last port.LastAttempt) error {
	if last.BuildErr == nil {
		return err
	}
	return fmt.Errorf("%w: %w", err, last)
}

// fetchFragment は run を 1 本起こして原稿断片を得る。agentID が空なら agent ごと create し、あれば同じ agent への follow-up run にする。
func (w *TextWriter) fetchFragment(ctx context.Context, agentID, brief string, last port.LastAttempt) (string, string, error) {
	agentID, runID, err := w.startRun(ctx, agentID, brief, last)
	if err != nil {
		return agentID, "", err
	}
	raw, err := w.streamResult(ctx, agentID, runID)
	return agentID, raw, err
}

// why: follow-up run は同一 agent の会話継続で、前回出力は Cursor 側が保持している。rejection 理由だけを送れば足りるので、brief も raw も再送しない。
func (w *TextWriter) startRun(ctx context.Context, agentID, brief string, last port.LastAttempt) (string, string, error) {
	if agentID == "" {
		return w.createAgent(ctx, brief)
	}
	runID, err := w.createRun(ctx, agentID, port.RejectionMiddleText+last.BuildErr.Error()+port.RejectionSuffixText)
	return agentID, runID, err
}

type createAgentRequest struct {
	Prompt promptText `json:"prompt"`
	Model  modelID    `json:"model"`
}

type promptText struct {
	Text string `json:"text"`
}

type modelID struct {
	ID string `json:"id"`
}

type createAgentResponse struct {
	Agent struct {
		ID string `json:"id"`
	} `json:"agent"`
	Run struct {
		ID string `json:"id"`
	} `json:"run"`
}

// why: createAgent は非 idempotent なので Do error・5xx・timeout でも再試行しない（二重 agent 回避）。
func (w *TextWriter) createAgent(ctx context.Context, brief string) (string, string, error) {
	body, err := json.Marshal(createAgentRequest{
		Prompt: promptText{Text: brief},
		Model:  modelID{ID: ModelID},
	})
	if err != nil {
		return "", "", infraErr("marshal_request", err)
	}

	raw, err := w.postJSON(ctx, APIBaseURL+AgentsPath, body, "create_status")
	if err != nil {
		return "", "", err
	}

	var parsed createAgentResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", "", infraErr("parse_create", err)
	}
	agentID := strings.TrimSpace(parsed.Agent.ID)
	runID := strings.TrimSpace(parsed.Run.ID)
	if agentID == "" || runID == "" {
		return "", "", infraErr("parse_create", fmt.Errorf("agent id or run id is missing"))
	}
	return agentID, runID, nil
}

type createRunRequest struct {
	Prompt promptText `json:"prompt"`
}

type createRunResponse struct {
	ID string `json:"id"`
}

// why: createRun も非 idempotent（同一 agent への追加会話）なので再試行しない。model は agent create 時に確定済みで、Create A Run は prompt だけを受け取る。
func (w *TextWriter) createRun(ctx context.Context, agentID, prompt string) (string, error) {
	body, err := json.Marshal(createRunRequest{Prompt: promptText{Text: prompt}})
	if err != nil {
		return "", infraErr("marshal_request", err)
	}

	raw, err := w.postJSON(ctx, fmt.Sprintf(RunsPathTemplate, APIBaseURL+AgentsPath, agentID), body, "create_run_status")
	if err != nil {
		return "", err
	}

	var parsed createRunResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", infraErr("parse_create_run", err)
	}
	runID := strings.TrimSpace(parsed.ID)
	if runID == "" {
		return "", infraErr("parse_create_run", fmt.Errorf("run id is missing"))
	}
	return runID, nil
}

func (w *TextWriter) postJSON(ctx context.Context, url string, body []byte, op string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, infraErr("build_request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(AuthorizationHeader, BearerTokenPrefix+w.apiKey)

	res, err := w.client.Do(req)
	if err != nil {
		return nil, passToFallback(infraErr("do", err))
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, passToFallback(infraErr("read_body", err))
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, createStatusError(op, res.StatusCode, raw)
	}
	return raw, nil
}

// createStatusError は create 系 POST の失敗 status を Infrastructure Error にし、fallback へ渡す status なら番兵で wrap する。
// why: System 失敗の切り分けに理由 body が要る。secret は Authorization header にしか載せないので、body を bounded で出しても credential は漏れない。
func createStatusError(op string, status int, raw []byte) error {
	err := infraErr(op, fmt.Errorf("status %d; response body: %s", status, httpdiag.BodySnippet(raw)))
	if shouldFallbackOnCreateStatus(status, raw) {
		return passToFallback(err)
	}
	return err
}

// why: create は非 idempotent で再試行しないので、429・5xx は待たずに fallback へ渡す。401・403 は API key・subscription の失効、
//
//	400 + usage_limit_exceeded は利用枠の喪失（run 34132953055。Decision generator-api-failure-retry-or-fallback）。
//	Cursor は error body の code を公式に定めていないので、parse せず文字列の存在だけを見る。
func shouldFallbackOnCreateStatus(status int, raw []byte) bool {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden,
		status == http.StatusTooManyRequests, status >= http.StatusInternalServerError:
		return true
	case status == http.StatusBadRequest:
		return bytes.Contains(raw, []byte(usageLimitExceededCode))
	default:
		return false
	}
}

// passToFallback は err を、呼び出し元が次の取得元へ切り替える番兵で wrap する。
func passToFallback(err error) error {
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
}

// streamRetryKind は stream 取得失敗の再試行方針。
type streamRetryKind int

const (
	// retryNone は再試行せず、そのまま error を返す（bug。4xx（429・401・403 除く）、run 終端 error、空 text、SSE 途中断、parse 失敗）。
	retryNone streamRetryKind = iota
	// retryOnce は Do error / idempotent GET の 5xx。+1 即再試行を 1 回だけ。使い切りは fallback へ渡す。
	retryOnce
	// retryRateLimited は Retry-After で回復が明示された 429。MaxAttempts まで待って再試行する。使い切りは fallback へ渡す。
	retryRateLimited
	// retryFallback は再試行せず fallback へ渡す。回復の明示が無い 429、API key の失効（401・403）。
	retryFallback
)

// streamRetryPolicy は失敗した stream 取得への対処。zero value は再試行しない。
type streamRetryPolicy struct {
	kind streamRetryKind
	wait time.Duration
}

// why: Cursor は 429 の機械可読 code を公式に定めていないので body を読まず、標準 header の Retry-After だけを回復の明示とする。
func classifyStreamStatus(status int, header http.Header) streamRetryPolicy {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return streamRetryPolicy{kind: retryFallback}
	case status == http.StatusTooManyRequests:
		if wait := retryAfter(header); wait > 0 {
			return streamRetryPolicy{kind: retryRateLimited, wait: wait}
		}
		return streamRetryPolicy{kind: retryFallback}
	case status >= http.StatusInternalServerError:
		return streamRetryPolicy{kind: retryOnce}
	default:
		return streamRetryPolicy{kind: retryNone}
	}
}

// giveUpError は再試行を終える時の error を返す。retryNone 以外は fallback へ渡す。
func giveUpError(kind streamRetryKind, err error) error {
	if kind == retryNone {
		return err
	}
	return passToFallback(err)
}

func (w *TextWriter) streamResult(ctx context.Context, agentID, runID string) (string, error) {
	url := fmt.Sprintf(StreamPathTemplate, APIBaseURL+AgentsPath, agentID, runID)

	onceRetried := false
	rateLimitedAttempts := 0
	for {
		text, policy, err := w.fetchStream(ctx, url)
		if err == nil {
			return text, nil
		}
		switch policy.kind {
		case retryOnce:
			if onceRetried {
				return "", giveUpError(policy.kind, err)
			}
			onceRetried = true
		case retryRateLimited:
			rateLimitedAttempts++
			if rateLimitedAttempts >= MaxAttempts {
				return "", giveUpError(policy.kind, err)
			}
			w.retry.Retry("stream_result", rateLimitedAttempts, MaxAttempts, err.Error())
			w.backoffSleepFn(ctx, policy.wait)
		default:
			return "", giveUpError(policy.kind, err)
		}
	}
}

// fetchStream は 1 回の GET を実行する。失敗時は error と、その失敗への再試行方針を返す。
func (w *TextWriter) fetchStream(ctx context.Context, url string) (string, streamRetryPolicy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", streamRetryPolicy{}, infraErr("build_request", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set(AuthorizationHeader, BearerTokenPrefix+w.apiKey)

	res, err := w.client.Do(req)
	if err != nil {
		return "", streamRetryPolicy{kind: retryOnce}, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return "", classifyStreamStatus(res.StatusCode, res.Header), infraErr("stream_status", fmt.Errorf("stream status %d", res.StatusCode))
	}

	text, err := parseResultText(res.Body)
	if err != nil {
		// why: SSE 途中断・parse 失敗は再 stream しない。run は既に終端しうるし、再 create は禁止（Decision 2026-09-03T17-03-33-feature-generator-cursor-cli-to-http-api）。
		return "", streamRetryPolicy{}, err
	}
	return text, streamRetryPolicy{}, nil
}

// why: delta-seconds 形式のみ尊重する。RFC 9110 の HTTP-date 形式は解釈せず backoff にフォールバックする（YAGNI）。
func retryAfter(header http.Header) time.Duration {
	v := strings.TrimSpace(header.Get("Retry-After"))
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(v)
	if err != nil || secs < 0 {
		return 0
	}
	d := time.Duration(secs) * time.Second
	if d > MaxRetryAfter {
		return MaxRetryAfter
	}
	return d
}

type resultEventData struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

func decodeResultEvent(data string) (string, error) {
	var parsed resultEventData
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return "", infraErr("parse_sse", err)
	}
	if parsed.Status != "" && parsed.Status != "FINISHED" {
		return "", infraErr("run_status", fmt.Errorf("run terminated with status %s", parsed.Status))
	}
	text := strings.TrimSpace(parsed.Text)
	if text == "" {
		return "", infraErr("empty_text", fmt.Errorf("result event has empty text"))
	}
	return text, nil
}

func parseResultText(body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, scanBufferInitialBytes), StreamBufferBytes)

	var eventName string
	var lastData string
	flush := func() (string, bool, error) {
		name, data := eventName, lastData
		eventName, lastData = "", ""
		if name != "result" || data == "" {
			return "", false, nil
		}
		text, err := decodeResultEvent(data)
		return text, err == nil, err
	}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			text, done, err := flush()
			if err != nil {
				return "", err
			}
			if done {
				return text, nil
			}
			continue
		}
		switch {
		case strings.HasPrefix(line, "event:"):
			eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			// why: SSE 仕様は data 複数行を許すが Cursor の result event は 1 行 JSON。最後の 1 行だけ見る。
			lastData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", infraErr("parse_sse", err)
	}
	text, done, err := flush()
	if err != nil {
		return "", err
	}
	if done {
		return text, nil
	}
	return "", infraErr("parse_sse", fmt.Errorf("stream ended without result event"))
}
