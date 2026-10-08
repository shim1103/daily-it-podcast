// Package gemini は Gemini Interactions API を使う朗読音声 SpeechSynthesizer Adapter を提供する。
// 音声合成の取得元が当面使えない時は port.ErrSourceExhausted を wrap して返し、UseCase が次の取得元へ渡す。
package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpdiag"
)

var _ port.SpeechSynthesizer = (*SpeechSynthesizer)(nil)

// Tuning は SpeechSynthesizer の待機系パラメータの注入値。
// ゼロ値 field は既定値（default*）へフォールバックする。
type Tuning struct {
	CallGap          time.Duration
	RetryBackoffBase time.Duration
	RetryBackoffMax  time.Duration
}

// SpeechSynthesizer は Gemini Interactions API で朗読音声を合成する Adapter。
type SpeechSynthesizer struct {
	client         *http.Client
	apiKey         string
	tier           Tier
	retry          port.RetryReporter
	backoffSleepFn func(time.Duration) // why: test の並列実行と共存するため package global に置かない
	lastCallAt     time.Time
	nowFn          func() time.Time
	// why: 待機系パラメータは rate 計測 test から注入で差し替えるため field にする（Decision 2026-09-03T14-46-00）。
	callGap          time.Duration
	retryBackoffBase time.Duration
	retryBackoffMax  time.Duration
}

// NewSpeechSynthesizer は Gemini TTS Adapter を返す。待機系パラメータは既定値。
//
// @require httpClient != nil。retry != nil（Composition Root の結線責務）。
// @ensure apiKey は x-goog-api-key header にだけ使い、保存元の知識は持たない。
func NewSpeechSynthesizer(httpClient *http.Client, apiKey string, tier Tier, retry port.RetryReporter) *SpeechSynthesizer {
	return NewSpeechSynthesizerWithTuning(httpClient, apiKey, tier, Tuning{}, retry)
}

// NewSpeechSynthesizerWithTuning は待機系パラメータを注入できる constructor。
// rate 計測（system && ratemeasure）専用の差し替え口。
//
// @require httpClient != nil。retry != nil（Composition Root の結線責務）。
// @ensure tuning のゼロ値 field は既定値（defaultCallGap / defaultRetryBackoffBase / defaultRetryBackoffMax）へフォールバックする。
// @ensure Tuning{} を渡した場合の挙動は NewSpeechSynthesizer と同一。
func NewSpeechSynthesizerWithTuning(httpClient *http.Client, apiKey string, tier Tier, tuning Tuning, retry port.RetryReporter) *SpeechSynthesizer {
	s := newSpeechSynthesizer(httpClient, apiKey, time.Sleep, retry)
	s.tier = tier
	s.callGap = firstNonZeroDuration(tuning.CallGap, defaultCallGap)
	s.retryBackoffBase = firstNonZeroDuration(tuning.RetryBackoffBase, defaultRetryBackoffBase)
	s.retryBackoffMax = firstNonZeroDuration(tuning.RetryBackoffMax, defaultRetryBackoffMax)
	return s
}

// SynthesizeAll は texts を順に朗読音声へ変換し、セグメント単位の WAV 列（結合しない）を返す。
//
// @require texts の各要素は trim 後に非空。
// @ensure 成功時は len(texts) と同数の非空・最小尺 WAV を返す。
// @ensure 失敗時もそれまでに合成できた分の audios（部分成功）を err と併せて返す。
// @ensure 呼び出し全体で Gemini 呼び出し合計を Tier ごとの上限（SynthesizeBudget / SynthesizeBudgetPaid）以内に抑える。1 セグメントは min(MaxAttempts, 残予算) 回まで。上限へ達した後のセグメントは即 error。
// @ensure 取得元が当面使えない失敗（認証断・利用枠喪失・回復の明示が無い 429・再試行の使い切り）は Tier を問わず port.ErrSourceExhausted を wrap して返す。それ以外（4xx・呼び出し上限超過など）は wrap せずそのまま返す（Decision generator-genai-api-failure-retry-or-fallback）。
func (s *SpeechSynthesizer) SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("synthesize", fmt.Errorf("client is nil"))
	}

	budget := s.synthesizeBudget()
	audios := make([]models.SpeechAudio, 0, len(texts))
	callsSpent := 0
	for i, text := range texts {
		remaining := budget - callsSpent
		if remaining <= 0 {
			return audios, budgetExhaustedError(i, len(texts), callsSpent, budget)
		}
		audio, used, err := s.synthesizeOne(ctx, text, min(MaxAttempts, remaining))
		callsSpent += used
		if err != nil {
			return audios, err
		}
		audios = append(audios, audio)
	}
	return audios, nil
}

func newSpeechSynthesizer(httpClient *http.Client, apiKey string, backoffSleepFn func(time.Duration), retry port.RetryReporter) *SpeechSynthesizer {
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	return &SpeechSynthesizer{
		client:           withCallTimeout(httpClient),
		apiKey:           apiKey,
		retry:            retry,
		backoffSleepFn:   backoffSleepFn,
		nowFn:            time.Now,
		callGap:          defaultCallGap,
		retryBackoffBase: defaultRetryBackoffBase,
		retryBackoffMax:  defaultRetryBackoffMax,
	}
}

func firstNonZeroDuration(v, fallback time.Duration) time.Duration {
	if v == 0 {
		return fallback
	}
	return v
}

// why: Composition が渡す *http.Client は全体 timeout を持たない。1 呼び出しの上限は
// vendor 固有制約なので Adapter が付け直し、呼び出し元の Client は変更しない。
func withCallTimeout(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		return nil
	}
	c := *httpClient
	c.Timeout = httpCallTimeout
	return &c
}

func (s *SpeechSynthesizer) synthesizeBudget() int {
	if s.tier == TierPaid {
		return SynthesizeBudgetPaid
	}
	return SynthesizeBudget
}

func budgetExhaustedError(segmentIndex, segmentCount, spent, budget int) error {
	return infraErr("synthesize_budget", fmt.Errorf(
		"gemini call budget exhausted at segment %d/%d: spent %d of %d", segmentIndex+1, segmentCount, spent, budget))
}

// what: 戻りの int は実際に消費した Gemini 呼び出し回数。SynthesizeAll が残予算の計算に使う。
func (s *SpeechSynthesizer) synthesizeOne(ctx context.Context, text string, maxAttempts int) (models.SpeechAudio, int, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return models.SpeechAudio{}, 0, infraErr("validate_text", fmt.Errorf("text is empty after trim"))
	}

	run := &segmentRun{maxAttempts: max(maxAttempts, 1)}
	for {
		s.waitCallGap()
		wav, failure := s.fetchWAV(ctx, trimmed)
		s.stampCallFinished()
		run.calls++
		if failure == nil {
			audio, err := toSpeechAudio(wav)
			return audio, run.calls, err
		}
		run.recordFailure(failure)
		if run.shouldGiveUp(failure) {
			return models.SpeechAudio{}, run.calls, wrapForFallback(failure.kind, failure.err)
		}
		s.reportRetryAndWait(run, failure)
	}
}

type segmentRun struct {
	maxAttempts int
	calls       int
	streak      failureStreak
}

func (r *segmentRun) recordFailure(failure *fetchFailure) {
	r.streak.record(failure.kind, failure.err)
}

func (r *segmentRun) shouldGiveUp(failure *fetchFailure) bool {
	return !failure.kind.retryable() || r.calls >= r.maxAttempts || r.streak.repeatsSameOp()
}

type failureStreak struct {
	last  error
	count int
}

func (f *failureStreak) record(kind speechFetchRetryKind, err error) {
	if kind.retryable() && sameGeminiOp(f.last, err) {
		f.count++
	} else {
		f.count = 1
	}
	f.last = err
}

func (f *failureStreak) repeatsSameOp() bool {
	return f.count >= maxConsecutiveSameOp
}

// why: adaptererror.Error は全 infra 共通型なので、Source も見ないと別 Adapter の同名 Op（"http_status" 等）が偶然一致しうる。
func sameGeminiOp(prev, cur error) bool {
	if prev == nil || cur == nil {
		return false
	}
	var prevErr, curErr *adaptererror.Error
	if !errors.As(prev, &prevErr) || !errors.As(cur, &curErr) {
		return false
	}
	return prevErr.Source == curErr.Source && prevErr.Op == curErr.Op
}

func toSpeechAudio(wav []byte) (models.SpeechAudio, error) {
	duration, err := wavDurationSec(wav)
	if err != nil {
		return models.SpeechAudio{}, infraErr("wav_duration", err)
	}
	return models.SpeechAudio{Content: wav, DurationSec: duration}, nil
}

func (s *SpeechSynthesizer) waitCallGap() {
	if s.lastCallAt.IsZero() {
		return
	}
	gap := s.effectiveCallGap()
	elapsed := s.nowFunc()().Sub(s.lastCallAt)
	if elapsed >= gap {
		return
	}
	s.sleepFunc()(gap - elapsed)
}

func (s *SpeechSynthesizer) stampCallFinished() {
	s.lastCallAt = s.nowFunc()()
}

func (s *SpeechSynthesizer) reportRetryAndWait(run *segmentRun, failure *fetchFailure) {
	s.retry.Retry("synthesize_speech", run.calls, run.maxAttempts, failure.err.Error())
	s.sleepFunc()(max(s.retryDelay(run.calls), failure.suggestedWait))
}

func (s *SpeechSynthesizer) effectiveCallGap() time.Duration {
	if s.callGap != 0 {
		return s.callGap
	}
	return defaultCallGap
}

func (s *SpeechSynthesizer) effectiveRetryBackoffBase() time.Duration {
	if s.retryBackoffBase != 0 {
		return s.retryBackoffBase
	}
	return defaultRetryBackoffBase
}

func (s *SpeechSynthesizer) effectiveRetryBackoffMax() time.Duration {
	if s.retryBackoffMax != 0 {
		return s.retryBackoffMax
	}
	return defaultRetryBackoffMax
}

func (s *SpeechSynthesizer) retryDelay(attempt int) time.Duration {
	// why: 公式は exponential。System の 429 対策で base を 60s・上限 3m にする。
	if attempt < 1 {
		attempt = 1
	}
	base := s.effectiveRetryBackoffBase()
	max := s.effectiveRetryBackoffMax()
	d := base << (attempt - 1)
	if d > max || d <= 0 {
		return max
	}
	return d
}

func (s *SpeechSynthesizer) parseRetryAfter(h http.Header) time.Duration {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	if raw == "" {
		return 0
	}
	sec, err := strconv.Atoi(raw)
	if err != nil || sec <= 0 {
		return 0
	}
	d := time.Duration(sec) * time.Second
	if max := s.effectiveRetryBackoffMax(); d > max {
		return max
	}
	return d
}

func (s *SpeechSynthesizer) sleepFunc() func(time.Duration) {
	if s.backoffSleepFn == nil {
		return time.Sleep
	}
	return s.backoffSleepFn
}

func (s *SpeechSynthesizer) nowFunc() func() time.Time {
	if s.nowFn == nil {
		return time.Now
	}
	return s.nowFn
}

type speechFetchRetryKind int

const (
	// what: 再試行も fallback もせず、error をそのまま返す。
	wavRetryNone speechFetchRetryKind = iota
	// what: 待って再試行する。同種 error の連続は synthesizeOne が打ち切る。
	wavRetryTransient
	// what: 回復が明示された 429。待って再試行する。
	wavRetryRateLimited
	// what: 再試行せず fallback へ渡す。
	wavRetryFallback
)

func (k speechFetchRetryKind) retryable() bool {
	return k == wavRetryTransient || k == wavRetryRateLimited
}

type fetchFailure struct {
	kind          speechFetchRetryKind
	err           error
	suggestedWait time.Duration
}

func failedFetch(kind speechFetchRetryKind, err error) *fetchFailure {
	return &fetchFailure{kind: kind, err: err}
}

// why: 再試行の使い切り（Transient・RateLimited）も Fallback と同じく fallback へ渡す。
func wrapForFallback(kind speechFetchRetryKind, err error) error {
	if kind == wavRetryNone {
		return err
	}
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
}

func (s *SpeechSynthesizer) fetchWAV(ctx context.Context, transcript string) ([]byte, *fetchFailure) {
	req, err := s.newSynthesizeRequest(ctx, transcript)
	if err != nil {
		return nil, failedFetch(wavRetryNone, err)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, failedFetch(wavRetryTransient, infraErr("do", err))
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, failedFetch(wavRetryTransient, infraErr("read_body", err))
	}

	if detectProhibitedContent(raw) {
		return nil, failedFetch(wavRetryNone, infraErr("prohibited_content", errors.New(prohibitedContentMarker)))
	}

	if res.StatusCode != http.StatusOK {
		kind, wait := s.classifyFailedStatus(res.StatusCode, res.Header, raw)
		return nil, &fetchFailure{
			kind:          kind,
			err:           infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw))),
			suggestedWait: wait,
		}
	}

	wav, err := decodeWAV(raw)
	if err != nil {
		// why: audio 欠落・極小音声は公式 Limitation の一過性劣化なので Transient にする。原因（finish_reason / safety / body 内 quota）を後から読めるよう、応答本文の snippet を error に載せる。
		return nil, failedFetch(wavRetryTransient, infraErr("decode_wav", fmt.Errorf("%w; response body: %s", err, httpdiag.BodySnippet(raw))))
	}
	return wav, nil
}

func (s *SpeechSynthesizer) newSynthesizeRequest(ctx context.Context, transcript string) (*http.Request, error) {
	body, err := json.Marshal(interactionRequest{
		Model:  ModelID,
		Input:  buildInput(transcript),
		Format: responseFormat{Type: "audio"},
		GenerationConfig: generationConfig{
			SpeechConfig: []speechConfig{{Voice: VoiceName}},
		},
	})
	if err != nil {
		return nil, infraErr("marshal_request", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, EndpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, infraErr("build_request", err)
	}
	req.Header.Set(geminiAPIKeyHeader, s.apiKey)
	return req, nil
}

func buildInput(transcript string) []interactionInput {
	return []interactionInput{
		{
			Type: "user_input",
			Content: []inputContent{
				{
					Type: "text",
					Text: transcript,
					Annotations: []inputAnnotation{
						{Type: "speech_metadata"},
					},
				},
			},
		},
	}
}

func detectProhibitedContent(body []byte) bool {
	return strings.Contains(string(body), prohibitedContentMarker)
}

// why: 429 の回復は response の message でなく、公式 error.code と Retry-After で見る。quota_exceeded は 400 でも返るので status を問わず見る（Decision generator-genai-api-failure-retry-or-fallback）。
func (s *SpeechSynthesizer) classifyFailedStatus(status int, header http.Header, raw []byte) (speechFetchRetryKind, time.Duration) {
	wait := s.parseRetryAfter(header)
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return wavRetryFallback, 0
	case (status == http.StatusBadRequest || status == http.StatusTooManyRequests) && quotaExceeded(raw):
		return wavRetryFallback, 0
	case status == http.StatusTooManyRequests:
		if wait > 0 || rateLimitExceeded(raw) {
			return wavRetryRateLimited, wait
		}
		return wavRetryFallback, 0
	case status >= http.StatusInternalServerError:
		return wavRetryTransient, wait
	default:
		return wavRetryNone, 0
	}
}

// why: provider が message の文言を変えても分類が壊れないよう、公式の error.code だけを読む。
func errorCode(raw []byte) string {
	var parsed errorResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ""
	}
	return parsed.Error.Code
}

func quotaExceeded(raw []byte) bool {
	return errorCode(raw) == errorCodeQuotaExceeded
}

func rateLimitExceeded(raw []byte) bool {
	return errorCode(raw) == errorCodeRateLimitExceeded
}

func decodeWAV(body []byte) ([]byte, error) {
	data, err := extractAudioBase64(body)
	if err != nil {
		return nil, err
	}
	wav, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	if err := validateAudioWAV(wav); err != nil {
		return nil, err
	}
	return wav, nil
}

func extractAudioBase64(body []byte) (string, error) {
	var parsed interactionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	data := firstAudioData(parsed)
	if data == "" {
		// why: struct 経由の parse では audio 欠落の原因が読めない。body のトップレベルキー一覧を添え、応答構造の想定違いを切り分ける。
		return "", fmt.Errorf("output audio is missing%s", topLevelKeysHint(body))
	}
	return data, nil
}

func validateAudioWAV(wav []byte) error {
	if len(wav) == 0 {
		return fmt.Errorf("output audio is empty")
	}
	duration, err := wavDurationSec(wav)
	if err != nil {
		return err
	}
	// why: HTTP 200 で返る極小音声は実質無音の一過性劣化。audio 欠落と同じく fetchWAV が Transient にする。
	if duration < minSpeechDurationSec {
		return fmt.Errorf("output audio is too short: %.2fs < %.1fs", duration, minSpeechDurationSec)
	}
	return nil
}

// why: 実 Interactions API の audio base64 は steps[].content[].data に入る。steps[0].content[0] へ決め打ちせず、空を飛ばして最初に見つかった data を採る。
func firstAudioData(parsed interactionResponse) string {
	for _, step := range parsed.Steps {
		for _, content := range step.Content {
			if trimmed := strings.TrimSpace(content.Data); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func topLevelKeysHint(body []byte) string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return ""
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Sprintf(" (top-level keys: [%s])", strings.Join(keys, ", "))
}

type interactionRequest struct {
	Model            string             `json:"model"`
	Input            []interactionInput `json:"input"`
	Format           responseFormat     `json:"response_format"`
	GenerationConfig generationConfig   `json:"generation_config"`
}

type interactionInput struct {
	Type    string         `json:"type"`
	Content []inputContent `json:"content"`
}

type inputContent struct {
	Type        string            `json:"type"`
	Text        string            `json:"text"`
	Annotations []inputAnnotation `json:"annotations,omitempty"`
}

type inputAnnotation struct {
	Type string `json:"type"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type generationConfig struct {
	SpeechConfig []speechConfig `json:"speech_config"`
}

type speechConfig struct {
	Voice string `json:"voice"`
}

type interactionResponse struct {
	Status string `json:"status"`
	Steps  []struct {
		Content []struct {
			Data string `json:"data"`
		} `json:"content"`
	} `json:"steps"`
}

type errorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}
