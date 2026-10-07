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
		return models.ManuscriptDraft{}, geminiErr("build_request", fmt.Errorf("client is nil"))
	}
	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return models.ManuscriptDraft{}, geminiErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	var last port.LastAttempt
	for attempt := 1; attempt <= TextWriterMaxAttempts; attempt++ {
		raw, err := w.generateContent(ctx, attemptBrief(trimmed, last))
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

func attemptBrief(brief string, last port.LastAttempt) string {
	if last.BuildErr == nil {
		return brief
	}
	return port.BuildRejectionBrief(brief, last.Raw, last.BuildErr.Error())
}

func withLastAttempt(err error, last port.LastAttempt) error {
	if last.BuildErr == nil {
		return err
	}
	return fmt.Errorf("%w: %w", err, last)
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

type fetchRetryKind int

const (
	// what: 再試行せず、error をそのまま返す（bug）。
	retryNone fetchRetryKind = iota
	// what: +1 即再試行を 1 回だけ行い、使い切りは fallback へ渡す。
	retryOnce
	// what: Retry-After で回復が明示された 429。MaxAttempts まで待って再試行し、使い切りは fallback へ渡す。
	retryRateLimited
	// what: 再試行せず fallback へ渡す。
	retryFallback
)

type fetchRetryPolicy struct {
	kind fetchRetryKind
	wait time.Duration
}

func terminalError(kind fetchRetryKind, err error) error {
	if kind == retryNone {
		return err
	}
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
}

// why: 429 は body の文言・code が provider 依存で変わりうるので読まず、標準 header の Retry-After だけを
// 回復の明示とする（Decision generator-genai-api-failure-retry-or-fallback）。
func classifyFailedStatus(status int, header http.Header) fetchRetryPolicy {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return fetchRetryPolicy{kind: retryFallback}
	case status == http.StatusTooManyRequests:
		if wait := retryAfter(header); wait > 0 {
			return fetchRetryPolicy{kind: retryRateLimited, wait: wait}
		}
		return fetchRetryPolicy{kind: retryFallback}
	case status >= http.StatusInternalServerError:
		return fetchRetryPolicy{kind: retryOnce}
	default:
		return fetchRetryPolicy{kind: retryNone}
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

// why: 再試行の方針は Decision generator-genai-api-failure-retry-or-fallback。この loop は fetchRetryKind だけを見る。
func (w *TextWriter) fetchWithRetry(ctx context.Context, url string, body []byte) (string, error) {
	retriedOnce := false
	rateLimitAttempt := 0
	for {
		text, policy, err := w.fetchOnce(ctx, url, body)
		if err == nil {
			return text, nil
		}
		switch policy.kind {
		case retryOnce:
			if retriedOnce {
				return "", terminalError(policy.kind, err)
			}
			retriedOnce = true
		case retryRateLimited:
			rateLimitAttempt++
			if rateLimitAttempt >= MaxAttempts {
				return "", terminalError(policy.kind, err)
			}
			w.retry.Retry("generate_content", rateLimitAttempt, MaxAttempts, err.Error())
			w.backoffSleepFn(ctx, policy.wait)
		default:
			return "", terminalError(policy.kind, err)
		}
	}
}

func (w *TextWriter) fetchOnce(ctx context.Context, url string, body []byte) (string, fetchRetryPolicy, error) {
	req, err := w.newRequest(ctx, url, body)
	if err != nil {
		return "", fetchRetryPolicy{}, err
	}

	res, err := w.client.Do(req)
	if err != nil {
		return "", fetchRetryPolicy{kind: retryOnce}, geminiErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(res.Body, ResponseBufferBytes))
	if err != nil {
		// why: body 読みは status 分岐の前。read 途中断（transient network 切断など）は 5xx と同じ一過性として扱う。
		return "", fetchRetryPolicy{kind: retryOnce}, geminiErr("read_body", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", classifyFailedStatus(res.StatusCode, res.Header), statusError(res.StatusCode, raw)
	}

	text, err := parseGeneratedText(raw)
	return text, fetchRetryPolicy{}, err
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
