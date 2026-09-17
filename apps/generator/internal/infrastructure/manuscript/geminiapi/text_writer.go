// Package geminiapi は Gemini generateContent を使う原稿 TextWriter Adapter を提供する。
// Cursor Cloud Agents の利用枠喪失時に manuscript UseCase が secondary として使う。
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
}

// NewTextWriter は Gemini generateContent 用 TextWriter を組み立てる。
//
// @require apiKey は Composition で検証済み。
// @ensure 戻りは port.TextWriter。apiKey は APIKeyHeader にだけ使う。
// @ensure client == nil のとき Write は geminiErr("build_request") を返す。
// @ensure httpClient には textWriterHTTPTimeout を付けた shallow copy を使う（引数の Client は変更しない）。
func NewTextWriter(client *http.Client, apiKey string, tier Tier) *TextWriter {
	return newTextWriter(withCallTimeout(client), apiKey, tier, ctxSleep)
}

// newTextWriter は backoff の sleep 関数を差し込める内部 constructor。
// why: NewTextWriter は本番の ctxSleep を固定し、test は待ちを観測する fake を渡す（cursorapi と同型）。
func newTextWriter(client *http.Client, apiKey string, tier Tier, backoffSleepFn func(context.Context, time.Duration)) *TextWriter {
	if backoffSleepFn == nil {
		backoffSleepFn = ctxSleep
	}
	return &TextWriter{client: client, apiKey: apiKey, tier: tier, backoffSleepFn: backoffSleepFn}
}

// withCallTimeout は httpClient の shallow copy に textWriterHTTPTimeout を付けて返す。
// why: 1 Write 呼び出し（invalid-draft retry を含む）の上限は vendor 固有制約なので、
//
//	timeout を持たない Composition 側の Client には頼らず Adapter が付け直す（gemini TTS の
//	withCallTimeout と同型）。
//
// @ensure httpClient == nil のときは nil を返す。
// @ensure 非 nil のときは textWriterHTTPTimeout を持つ別 *http.Client（引数は変更しない）。
func withCallTimeout(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		return nil
	}
	c := *httpClient
	c.Timeout = textWriterHTTPTimeout
	return &c
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

// Write は brief から invalid-draft retry（最大 TextWriterMaxAttempts 回）で valid な ManuscriptDraft を得る。
//
// @require brief は trim 後に非空。buildFn は非 nil。
// @ensure 成功時は buildFn が返す非 nil な models.ManuscriptDraft を返す。
// @ensure generateContent 自体が retry しない error を返してループを抜けるとき、直前 attempt で
//
//	buildFn が invalid と判定していれば（lastRaw/lastBuildErr が揃っていれば）、その情報を
//	port.LastAttempt として fmt.Errorf("%w: %w", err, port.LastAttempt{...}) の chain に含めて返す。
//	直前 attempt が無ければ error をそのまま透過する。
//
// @ensure buildFn が TextWriterMaxAttempts 回とも error を返したら、最後の error を
//
//	fmt.Errorf("%w: %w: %w", port.ErrDraftRejected, lastErr, port.LastAttempt{...}) で wrap して返す。
//	2 回目以降の試行は port.BuildRejectionBrief で前回の raw response と rejection 理由を
//	区別可能な形で brief へ埋め込む。
//
// @invariant generateContent は idempotent（同 body は同じ生成試行・副作用なし）。client.Do error / 5xx を 1 回、429 を MaxAttempts まで backoff で再試行し、429 の使い切りは Tier に関係なく port.ErrSourceExhausted を wrap する。401 / 403 / その他 4xx、finishReason が STOP 以外、空 text、parse 失敗は再試行しない。secret 実値を error へ出さない。model は ModelID 固定。
func (w *TextWriter) Write(ctx context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	if w == nil || w.client == nil {
		return models.ManuscriptDraft{}, geminiErr("build_request", fmt.Errorf("client is nil"))
	}
	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return models.ManuscriptDraft{}, geminiErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	attemptBrief := trimmed
	var lastErr error
	var lastRaw string
	var lastBuildErr error
	for attempt := 1; attempt <= TextWriterMaxAttempts; attempt++ {
		raw, err := w.generateContent(ctx, attemptBrief)
		if err != nil {
			if lastBuildErr != nil {
				err = fmt.Errorf("%w: %w", err, port.LastAttempt{Raw: lastRaw, BuildErr: lastBuildErr})
			}
			return models.ManuscriptDraft{}, err
		}
		draft, err := buildFn(raw)
		if err == nil {
			return draft, nil
		}
		lastErr = err
		lastRaw = raw
		lastBuildErr = err
		attemptBrief = port.BuildRejectionBrief(trimmed, lastRaw, lastBuildErr.Error())
	}
	return models.ManuscriptDraft{}, fmt.Errorf("%w: %w: %w", port.ErrDraftRejected, lastErr, port.LastAttempt{Raw: lastRaw, BuildErr: lastBuildErr})
}

type generateContentRequest struct {
	Contents []requestContent `json:"contents"`
	Tools    []requestTool    `json:"tools,omitempty"`
}

type requestContent struct {
	Parts []requestPart `json:"parts"`
}

type requestPart struct {
	Text string `json:"text"`
}

// requestTool は generateContent の tools 配列 1 要素。
// why: Gemini API v1beta generateContent の tools は {"type": "..."} 形式ではなく、tool 種別名を
//
//	key に持つ object（値は空 object）。ai.google.dev/gemini-api/docs/generate-content/url-context・
//	generate-content/google-search の REST curl 例で確認した実際の schema に合わせる
//	（{"type": "url_context"} 形式は新しい Interactions API 専用で generateContent には使えない）。
type requestTool struct {
	URLContext   *struct{} `json:"url_context,omitempty"`
	GoogleSearch *struct{} `json:"google_search,omitempty"`
}

// generateContentTools は毎回送る tools 配列。モデルが自律的に検索するか、prompt 中の URL を
// 深掘りするかを判断できるよう url_context と google_search の両方を常に有効にする
// （Decision: HackerNews 等の source URL は既に prompt にあるので、深掘りの要否判断はモデルに委ねる）。
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
	// retryNone は再試行しない（4xx（429 除く）、finishReason≠STOP、空 text、parse 失敗）。
	retryNone fetchRetryKind = iota
	// retryTransientOnce は Do error / idempotent POST の 5xx。+1 即再試行を 1 回だけ。
	retryTransientOnce
	// retryRateLimited は 429。MaxAttempts まで backoff で再試行する。
	retryRateLimited
)

// generateContent は generateContent を retry 方針に従って叩き、連結 trim 済み text を返す。
// why: Decision §7。Do error / 5xx は 1 回だけ（generateContent は idempotent）、429 を MaxAttempts まで backoff。
func (w *TextWriter) generateContent(ctx context.Context, brief string) (string, error) {
	body, err := json.Marshal(generateContentRequest{
		Contents: []requestContent{{Parts: []requestPart{{Text: brief}}}},
		Tools:    generateContentTools,
	})
	if err != nil {
		return "", geminiErr("marshal_request", err)
	}
	url := fmt.Sprintf(EndpointURLTemplate, ModelID)

	transientUsed := false
	rateLimitAttempt := 0
	for {
		text, kind, wait, err := w.fetchOnce(ctx, url, body)
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
				return "", fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
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

// backoffDelay は 429 再試行の待ち時間（1s, 2s, 4s…）。cursorapi.backoffDelay と同型。
func backoffDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return time.Second << (attempt - 1)
}

// fetchOnce は 1 回の POST を実行し、(断片, 再試行方針, 追加待ち, error) を返す。
func (w *TextWriter) fetchOnce(ctx context.Context, url string, body []byte) (string, fetchRetryKind, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", retryNone, 0, geminiErr("build_request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(APIKeyHeader, w.apiKey)

	res, err := w.client.Do(req)
	if err != nil {
		return "", retryTransientOnce, 0, geminiErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(res.Body, ResponseBufferBytes))
	if err != nil {
		// why: body 読みは status 分岐の前。read 途中断（transient network 切断など）は 5xx と同じ一過性として 1 回だけ再試行する。
		return "", retryTransientOnce, 0, geminiErr("read_body", err)
	}

	// why: secret は APIKeyHeader（x-goog-api-key）にしか載せないので、応答 body 全体を診断へ出しても credential は漏れない。
	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return "", retryRateLimited, retryAfter(res.Header), geminiErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	case res.StatusCode >= 500:
		return "", retryTransientOnce, 0, geminiErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	case res.StatusCode != http.StatusOK:
		return "", retryNone, 0, geminiErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	}

	text, err := parseGeneratedText(raw)
	if err != nil {
		return "", retryNone, 0, err
	}
	return text, retryNone, 0, nil
}

// retryAfter は Retry-After header を待ち時間へ変換する。無ければ 0、MaxRetryAfter でクランプ。
// why: cursorapi.retryAfter と同型。delta-seconds 形式のみ尊重し、RFC 9110 の HTTP-date 形式は
//
//	解釈せず backoff にフォールバックする（YAGNI。Gemini の 429 は delta-seconds で返る）。
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

// parseGeneratedText は generateContent 応答を parse し、candidates[0] の parts.text を連結 trim して返す。
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
