// Package geminiapi は Gemini generateContent を使う原稿 TextWriter Adapter を提供する。
// Cursor Cloud Agents が当面使えない時に manuscript UseCase が secondary として使う。
package geminiapi

import (
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

// TextWriter は Gemini generateContent を叩く原稿 Adapter。
type TextWriter struct {
	client         *http.Client
	apiKey         string
	tier           Tier
	backoffSleepFn func(context.Context, time.Duration) // why: test の並列実行と共存するため package global に置かない
	retry          port.RetryReporter
}

// NewTextWriter は Gemini generateContent 用 TextWriter を組み立てる。
//
// @require apiKey は Composition で検証済み。retry は非 nil（Composition Root の結線責務）。
// @ensure apiKey は APIKeyHeader にだけ使う。
// @ensure client が nil のとき Write は geminiErr("build_request") を返す。
// @ensure client は textWriterHTTPTimeout を付けた shallow copy として保持する（引数の Client は変更しない）。
func NewTextWriter(client *http.Client, apiKey string, tier Tier, retry port.RetryReporter) *TextWriter {
	return newTextWriter(withCallTimeout(client), apiKey, tier, ctxSleep, retry)
}

// Write は brief から valid な ManuscriptDraft を得る。
//
// @require brief は trim 後に非空。buildFn は非 nil。
// @ensure 成功時は buildFn が返した非 nil な models.ManuscriptDraft を返す。
// @ensure buildFn が invalid を返したら、前回の raw response と理由を埋めた brief（port.BuildRejectionBrief）で
// 最大 TextWriterMaxAttempts 回まで取り直す。次 attempt が残る時は w.retry へ
// Retry("write_manuscript_draft", attempt, TextWriterMaxAttempts, ...) を通知する。
// @ensure TextWriterMaxAttempts 回とも invalid なら、
// fmt.Errorf("%w: %w: %w", port.ErrDraftRejected, 最後の buildFn error, port.LastAttempt{...}) を返す。
// @ensure generateContent が error を返した時、それ以前の attempt で buildFn が invalid を返していれば、
// その raw と理由を port.LastAttempt として error の chain へ含める。無ければ error をそのまま返す。
// @ensure 取得元が当面使えない時は port.ErrSourceExhausted を wrap して返す。対象は、Do error・body 読み取り途中断・5xx が
// 1 回の再試行後も続く時、429 に解釈できる Retry-After が無い時、Retry-After 付き 429 が MaxAttempts 回続く時、401・403
// （API key の失効）の時。Tier は問わない。
// @ensure 上記以外の 4xx、finishReason が STOP 以外、空 text、parse 失敗は再試行も port.ErrSourceExhausted の wrap もしない。
// @invariant generateContent は idempotent で、再試行は副作用を持たない。secret 実値を error へ出さない。
func (w *TextWriter) Write(ctx context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	if w == nil || w.client == nil {
		return models.ManuscriptDraft{}, geminiErr("build_request", fmt.Errorf("client is nil"))
	}
	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return models.ManuscriptDraft{}, geminiErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	attemptBrief := trimmed
	var lastRaw string
	var lastBuildErr error
	for attempt := 1; attempt <= TextWriterMaxAttempts; attempt++ {
		raw, err := w.generateContent(ctx, attemptBrief)
		if err != nil {
			return models.ManuscriptDraft{}, withLastAttempt(err, lastRaw, lastBuildErr)
		}
		draft, buildErr := buildFn(raw)
		if buildErr == nil {
			return draft, nil
		}
		lastRaw, lastBuildErr = raw, buildErr
		if attempt < TextWriterMaxAttempts {
			w.retry.Retry("write_manuscript_draft", attempt, TextWriterMaxAttempts, lastBuildErr.Error())
			attemptBrief = port.BuildRejectionBrief(trimmed, lastRaw, lastBuildErr.Error())
		}
	}
	return models.ManuscriptDraft{}, fmt.Errorf("%w: %w: %w", port.ErrDraftRejected, lastBuildErr, port.LastAttempt{Raw: lastRaw, BuildErr: lastBuildErr})
}

// why: NewTextWriter は本番の ctxSleep を固定し、test は待ちを観測する fake を渡せるようにする。
func newTextWriter(client *http.Client, apiKey string, tier Tier, backoffSleepFn func(context.Context, time.Duration), retry port.RetryReporter) *TextWriter {
	if backoffSleepFn == nil {
		backoffSleepFn = ctxSleep
	}
	return &TextWriter{client: client, apiKey: apiKey, tier: tier, backoffSleepFn: backoffSleepFn, retry: retry}
}

// why: 1 Write（invalid-draft retry を含む）の上限は vendor 固有の制約なので、
// timeout を持たない Composition 側の Client に頼らず Adapter が付け直す。
func withCallTimeout(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		return nil
	}
	c := *httpClient
	c.Timeout = textWriterHTTPTimeout
	return &c
}

func ctxSleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func withLastAttempt(err error, raw string, buildErr error) error {
	if buildErr == nil {
		return err
	}
	return fmt.Errorf("%w: %w", err, port.LastAttempt{Raw: raw, BuildErr: buildErr})
}

type generateContentRequest struct {
	Contents         []requestContent  `json:"contents"`
	Tools            []requestTool     `json:"tools,omitempty"`
	GenerationConfig *generationConfig `json:"generationConfig,omitempty"`
}

type requestContent struct {
	Parts []requestPart `json:"parts"`
}

type requestPart struct {
	Text string `json:"text"`
}

// why: generateContent の tools は {"type": ...} 形式ではなく、tool 種別名を key に持つ object（値は空 object）。
// {"type": "url_context"} 形式は Interactions API 専用で generateContent には使えない
// （ai.google.dev の url-context・google-search の REST curl 例で確認した）。
type requestTool struct {
	URLContext   *struct{} `json:"url_context,omitempty"`
	GoogleSearch *struct{} `json:"google_search,omitempty"`
}

// why: source URL は既に prompt にあるので、深掘りするか検索するかはモデルが判断できるよう両方を常に有効にする。
var generateContentTools = []requestTool{
	{URLContext: &struct{}{}},
	{GoogleSearch: &struct{}{}},
}

type generateContentResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content      candidateContent `json:"content"`
	FinishReason string           `json:"finishReason"`
}

type candidateContent struct {
	Parts []contentPart `json:"parts"`
}

type contentPart struct {
	Text string `json:"text"`
}

// fetchRetryKind は generateContent 取得失敗の再試行方針。
type fetchRetryKind int

const (
	// retryNone は再試行せず、そのまま error を返す（bug。4xx（429・401・403 除く）、finishReason≠STOP、空 text、parse 失敗）。
	retryNone fetchRetryKind = iota
	// retryOnce は Do error / body 読み取り途中断 / 5xx。+1 即再試行を 1 回だけ。使い切りは fallback へ渡す。
	retryOnce
	// retryRateLimited は Retry-After で回復が明示された 429。MaxAttempts まで待って再試行する。使い切りは fallback へ渡す。
	retryRateLimited
	// retryFallback は再試行せず fallback へ渡す。回復の明示が無い 429、API key の失効（401・403）。
	retryFallback
)

func terminalError(kind fetchRetryKind, err error) error {
	if kind == retryNone {
		return err
	}
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
}

// why: 429 は body の文言・code が provider 依存で変わりうるので読まず、標準 header の Retry-After だけを
// 回復の明示とする（Decision generator-api-failure-retry-or-fallback）。
func classifyFailedStatus(status int, header http.Header) (fetchRetryKind, time.Duration) {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return retryFallback, 0
	case status == http.StatusTooManyRequests:
		if wait := retryAfter(header); wait > 0 {
			return retryRateLimited, wait
		}
		return retryFallback, 0
	case status >= http.StatusInternalServerError:
		return retryOnce, 0
	default:
		return retryNone, 0
	}
}

func (w *TextWriter) generateContent(ctx context.Context, brief string) (string, error) {
	body, err := encodeGenerateContentRequest(brief)
	if err != nil {
		return "", err
	}
	return w.fetchWithRetry(ctx, fmt.Sprintf(EndpointURLTemplate, ModelID), body)
}

func encodeGenerateContentRequest(brief string) ([]byte, error) {
	body, err := json.Marshal(generateContentRequest{
		Contents:         []requestContent{{Parts: []requestPart{{Text: brief}}}},
		Tools:            generateContentTools,
		GenerationConfig: &writerOutputGenerationConfig,
	})
	if err != nil {
		return nil, geminiErr("marshal_request", err)
	}
	return body, nil
}

// why: 再試行の方針は Decision generator-api-failure-retry-or-fallback。この loop は fetchRetryKind だけを見る。
func (w *TextWriter) fetchWithRetry(ctx context.Context, url string, body []byte) (string, error) {
	retriedOnce := false
	rateLimitAttempt := 0
	for {
		text, kind, wait, err := w.fetchOnce(ctx, url, body)
		if err == nil {
			return text, nil
		}
		switch kind {
		case retryOnce:
			if retriedOnce {
				return "", terminalError(kind, err)
			}
			retriedOnce = true
		case retryRateLimited:
			rateLimitAttempt++
			if rateLimitAttempt >= MaxAttempts {
				return "", terminalError(kind, err)
			}
			w.retry.Retry("generate_content", rateLimitAttempt, MaxAttempts, err.Error())
			w.backoffSleepFn(ctx, wait)
		default:
			return "", terminalError(kind, err)
		}
	}
}

func (w *TextWriter) fetchOnce(ctx context.Context, url string, body []byte) (text string, kind fetchRetryKind, wait time.Duration, err error) {
	req, err := w.newRequest(ctx, url, body)
	if err != nil {
		return "", retryNone, 0, err
	}

	res, err := w.client.Do(req)
	if err != nil {
		return "", retryOnce, 0, geminiErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(res.Body, ResponseBufferBytes))
	if err != nil {
		// why: body 読みは status 分岐の前。read 途中断（transient network 切断など）は 5xx と同じ一過性として扱う。
		return "", retryOnce, 0, geminiErr("read_body", err)
	}

	if res.StatusCode != http.StatusOK {
		kind, wait = classifyFailedStatus(res.StatusCode, res.Header)
		return "", kind, wait, statusError(res.StatusCode, raw)
	}

	text, err = parseGeneratedText(raw)
	return text, retryNone, 0, err
}

func (w *TextWriter) newRequest(ctx context.Context, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, geminiErr("build_request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(APIKeyHeader, w.apiKey)
	return req, nil
}

// why: secret は APIKeyHeader（x-goog-api-key）にしか載せないので、応答 body 全体を診断へ出しても credential は漏れない。
func statusError(status int, raw []byte) error {
	return geminiErr("http_status", fmt.Errorf("status %d; response body: %s", status, httpdiag.BodySnippet(raw)))
}

// why: delta-seconds 形式だけを回復の明示とする。HTTP-date 形式は解釈せず「明示なし」にする
// （YAGNI。Gemini の 429 は delta-seconds で返る）。
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

func parseGeneratedText(raw []byte) (string, error) {
	var parsed generateContentResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", geminiErr("parse_response", err)
	}
	if len(parsed.Candidates) == 0 {
		return "", geminiErr("parse_response", fmt.Errorf("response has no candidates"))
	}
	// why: request は candidateCount 未指定なので generateContent は 1 候補しか返さない。[0] で足りる。
	first := parsed.Candidates[0]
	if fr := first.FinishReason; fr != "" && fr != "STOP" {
		// why: finishReason の値は生成の打ち切り理由で secret ではない。そのまま出す。
		return "", geminiErr("finish_reason", fmt.Errorf("finish reason %s", fr))
	}
	var b strings.Builder
	for _, part := range first.Content.Parts {
		b.WriteString(part.Text)
	}
	text := strings.TrimSpace(b.String())
	if text == "" {
		return "", geminiErr("empty_text", fmt.Errorf("candidate has empty text"))
	}
	return text, nil
}
