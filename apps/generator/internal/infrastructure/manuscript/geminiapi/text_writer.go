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
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpdiag"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/writerretry"
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
	return newTextWriter(withCallTimeout(client), apiKey, tier, writerretry.SleepContext, retry)
}

// Write は brief から原稿断片を得て、buildFn で ManuscriptDraft へ解釈する。
//
// @require brief は trim 後に非空。buildFn は非 nil。
// @ensure 成功時は buildFn が返す非 nil な models.ManuscriptDraft を返す。
// @ensure buildFn が invalid と判定した後の取得 error には port.LastAttempt を chain に含める。
// @ensure buildFn が writerretry.MaxDraftAttempts 回とも error を返したら、port.ErrDraftRejected・最後の error・port.LastAttempt を wrap して返す。
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

	return writerretry.DraftRun{
		Retry: w.retry,
		Build: buildFn,
		Fetch: func(ctx context.Context, last port.LastAttempt) (string, error) {
			return w.generateContent(ctx, attemptBrief(trimmed, last))
		},
	}.Run(ctx)
}

// why: NewTextWriter は本番の writerretry.SleepContext を固定し、test は待ちを観測する fake を渡せるようにする。
func newTextWriter(client *http.Client, apiKey string, tier Tier, backoffSleepFn func(context.Context, time.Duration), retry port.RetryReporter) *TextWriter {
	if backoffSleepFn == nil {
		backoffSleepFn = writerretry.SleepContext
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

func attemptBrief(brief string, last port.LastAttempt) string {
	if last.BuildErr == nil {
		return brief
	}
	return port.BuildRejectionBrief(brief, last.Raw, last.BuildErr.Error())
}

func (w *TextWriter) generateContent(ctx context.Context, brief string) (string, error) {
	body, err := encodeGenerateContentRequest(brief)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(EndpointURLTemplate, ModelID)
	return writerretry.Run(ctx, w.retryConfig(), func(ctx context.Context) (string, writerretry.Policy, error) {
		return w.fetchOnce(ctx, url, body)
	})
}

func (w *TextWriter) retryConfig() writerretry.Config {
	return writerretry.Config{Step: "generate_content", Retry: w.retry, Sleep: w.backoffSleepFn}
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

func (w *TextWriter) fetchOnce(ctx context.Context, url string, body []byte) (string, writerretry.Policy, error) {
	req, err := w.newRequest(ctx, url, body)
	if err != nil {
		return "", writerretry.Policy{}, err
	}

	res, err := w.client.Do(req)
	if err != nil {
		return "", writerretry.Policy{Kind: writerretry.Once}, geminiErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(res.Body, ResponseBufferBytes))
	if err != nil {
		// why: body 読みは status 分岐の前。read 途中断（transient network 切断など）は 5xx と同じ一過性として扱う。
		return "", writerretry.Policy{Kind: writerretry.Once}, geminiErr("read_body", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", writerretry.ClassifyStatus(res.StatusCode, res.Header), statusError(res.StatusCode, raw)
	}

	text, err := parseGeneratedText(raw)
	return text, writerretry.Policy{}, err
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
