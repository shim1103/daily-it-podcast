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
)

var _ port.TextWriter = (*TextWriter)(nil)

// TextWriter は Cursor Cloud Agents API を使う原稿 Adapter。
type TextWriter struct {
	client         *http.Client
	apiKey         string
	backoffSleepFn func(context.Context, time.Duration) // why: test の並列実行と共存するため package global に置かない
}

// NewTextWriter は Cursor Cloud Agents API 用 TextWriter を組み立てる。
//
// @require apiKey は Composition で検証済み。
// @ensure 戻りは port.TextWriter。apiKey は Authorization: Bearer header にだけ使う。
// @ensure client == nil のとき Write は infraErr("build_request") を返す。
func NewTextWriter(client *http.Client, apiKey string) *TextWriter {
	return newTextWriterForTest(client, apiKey, ctxSleep)
}

func newTextWriterForTest(client *http.Client, apiKey string, backoffSleepFn func(context.Context, time.Duration)) *TextWriter {
	if backoffSleepFn == nil {
		backoffSleepFn = ctxSleep
	}
	return &TextWriter{client: client, apiKey: apiKey, backoffSleepFn: backoffSleepFn}
}

// ctxSleep は ctx が先に切れたらそちらを優先して待ちを中断する。
func ctxSleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// Write は brief から Cloud Agents（毎回 create → SSE 終端 result.text）で原稿断片を得る。
//
// @require brief は trim 後に非空。
// @ensure 成功時は非空 text 断片を返す。失敗時は *cursorapi.Error、断片は空。
// @invariant create（POST /v1/agents）は no-repo・非 retry。SSE 取得は idempotent GET として Do error / 5xx を 1 回、429 を MaxAttempts まで backoff で再試行する。SSE 途中断は再 stream しない。secret 実値は error へ出さない。
func (w *TextWriter) Write(ctx context.Context, brief string) (string, error) {
	if w == nil || w.client == nil {
		return "", infraErr("build_request", fmt.Errorf("client is nil"))
	}
	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return "", infraErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	agentID, runID, err := w.createAgent(ctx, trimmed)
	if err != nil {
		return "", err
	}
	return w.streamResult(ctx, agentID, runID)
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

// createAgent は no-repo で agent を create し、agentId と runId を返す。
// why: 非 idempotent なので Do error / 5xx / timeout でも再試行しない（二重 agent 回避）。
func (w *TextWriter) createAgent(ctx context.Context, brief string) (string, string, error) {
	body, err := json.Marshal(createAgentRequest{
		Prompt: promptText{Text: brief},
		Model:  modelID{ID: ModelID},
	})
	if err != nil {
		return "", "", infraErr("marshal_request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIBaseURL+AgentsPath, bytes.NewReader(body))
	if err != nil {
		return "", "", infraErr("build_request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(AuthorizationHeader, BearerTokenPrefix+w.apiKey)

	res, err := w.client.Do(req)
	if err != nil {
		return "", "", infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", infraErr("read_body", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", "", infraErr("create_status", fmt.Errorf("create status %d", res.StatusCode))
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

// streamRetryKind は stream 取得失敗の再試行方針。
type streamRetryKind int

const (
	// retryNone は再試行しない（4xx（429 除く）、run 終端 error、空 text、SSE 途中断、parse 失敗）。
	retryNone streamRetryKind = iota
	// retryTransientOnce は Do error / idempotent GET の 5xx。+1 即再試行を 1 回だけ。
	retryTransientOnce
	// retryRateLimited は 429。MaxAttempts まで backoff で再試行する。
	retryRateLimited
)

// streamResult は run の SSE を読み、終端 result event の text を断片として返す。
// why: Decision §5。Do error / 5xx は 1 回だけ、429 は MaxAttempts まで backoff。
func (w *TextWriter) streamResult(ctx context.Context, agentID, runID string) (string, error) {
	url := fmt.Sprintf(StreamPathTemplate, APIBaseURL+AgentsPath, agentID, runID)

	transientUsed := false
	rateLimitAttempt := 0
	for {
		text, kind, wait, err := w.fetchStream(ctx, url)
		if err == nil {
			return text, nil
		}
		switch kind {
		case retryTransientOnce:
			if transientUsed {
				return "", err
			}
			transientUsed = true
		case retryRateLimited:
			rateLimitAttempt++
			if rateLimitAttempt >= MaxAttempts {
				return "", err
			}
			if wait <= 0 {
				wait = backoffDelay(rateLimitAttempt)
			}
			w.backoffSleepFn(ctx, wait)
		default:
			return "", err
		}
	}
}

// backoffDelay は 429 再試行の待ち時間（1s, 2s, 4s…）。
// why: Cursor docs が 429 に exponential backoff を勧める。gemini の retryDelay と同型。
func backoffDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return time.Second << (attempt - 1)
}

// fetchStream は 1 回の GET を実行し、(断片, 再試行方針, 追加待ち, error) を返す。
func (w *TextWriter) fetchStream(ctx context.Context, url string) (string, streamRetryKind, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", retryNone, 0, infraErr("build_request", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set(AuthorizationHeader, BearerTokenPrefix+w.apiKey)

	res, err := w.client.Do(req)
	if err != nil {
		return "", retryTransientOnce, 0, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return "", retryRateLimited, retryAfter(res.Header), infraErr("stream_status", fmt.Errorf("stream status %d", res.StatusCode))
	case res.StatusCode >= 500:
		return "", retryTransientOnce, 0, infraErr("stream_status", fmt.Errorf("stream status %d", res.StatusCode))
	case res.StatusCode != http.StatusOK:
		return "", retryNone, 0, infraErr("stream_status", fmt.Errorf("stream status %d", res.StatusCode))
	}

	text, err := parseResultText(res.Body)
	if err != nil {
		// why: SSE 途中断・parse 失敗は再 stream しない。run は既に終端しうるし、再 create は Decision §5 で禁止。
		return "", retryNone, 0, err
	}
	return text, retryNone, 0, nil
}

// retryAfter は Retry-After header を待ち時間へ変換する。無ければ 0、MaxRetryAfter でクランプ。
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

// parseResultText は text/event-stream を走査し、終端 result event の text を返す。
func parseResultText(body io.Reader) (string, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), StreamBufferBytes)

	var eventName string
	var lastData string
	flush := func() (string, bool, error) {
		name, data := eventName, lastData
		eventName, lastData = "", ""
		if name != "result" || data == "" {
			return "", false, nil
		}
		var parsed resultEventData
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			return "", false, infraErr("parse_sse", err)
		}
		if parsed.Status != "" && parsed.Status != "FINISHED" {
			return "", false, infraErr("run_status", fmt.Errorf("run terminated with status %s", parsed.Status))
		}
		text := strings.TrimSpace(parsed.Text)
		if text == "" {
			return "", false, infraErr("empty_text", fmt.Errorf("result event has empty text"))
		}
		return text, true, nil
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
