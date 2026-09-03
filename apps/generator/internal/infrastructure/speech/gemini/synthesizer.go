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
	"strconv"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

const geminiAPIKeyHeader = "x-goog-api-key"

var _ port.SpeechSynthesizer = (*SpeechSynthesizer)(nil)

// synthesizeBudget と maxAttempts の関係:
//   - MaxAttempts   : 1 セグメントが連続で消費してよい上限（暴走ガード）。
//   - SynthesizeBudget: 1 度の SynthesizeAll 全体で許す合計上限（RPD ガード）。
//     各セグメントは min(MaxAttempts, 残予算) 回まで。

type SpeechSynthesizer struct {
	client         *http.Client
	apiKey         string
	backoffSleepFn func(time.Duration) // why: test の並列実行と共存するため package global に置かない
	lastCallAt     time.Time
	nowFn          func() time.Time
	// why: 429 頻度と総所要のトレードオフを実測で詰めるため、待機系パラメータを field 化して
	//      rate 計測 test から注入で差し替える（Decision 2026-09-03T14-46-00）。
	//      既定 constructor は default* const を入れるので挙動は不変。MaxAttempts は注入対象外。
	callGap          time.Duration
	retryBackoffBase time.Duration
	retryBackoffMax  time.Duration
}

// Tuning は SpeechSynthesizer の待機系パラメータの注入値。
// ゼロ値 field は既定値（default*）へフォールバックする。
type Tuning struct {
	CallGap          time.Duration
	RetryBackoffBase time.Duration
	RetryBackoffMax  time.Duration
}

// NewSpeechSynthesizer は Gemini TTS Adapter を返す。待機系パラメータは既定値。
//
// @require httpClient != nil
// @ensure apiKey は x-goog-api-key header にだけ使い、保存元の知識は持たない。
func NewSpeechSynthesizer(httpClient *http.Client, apiKey string) *SpeechSynthesizer {
	return NewSpeechSynthesizerWithTuning(httpClient, apiKey, Tuning{})
}

// NewSpeechSynthesizerWithTuning は待機系パラメータを注入できる constructor。
// rate 計測（system && ratemeasure）専用の差し替え口。
//
// @require httpClient != nil
// @ensure tuning のゼロ値 field は既定値（defaultCallGap / defaultRetryBackoffBase / defaultRetryBackoffMax）へフォールバックする。
// @ensure Tuning{} を渡した場合の挙動は NewSpeechSynthesizer と同一。
func NewSpeechSynthesizerWithTuning(httpClient *http.Client, apiKey string, tuning Tuning) *SpeechSynthesizer {
	s := newSpeechSynthesizerForTest(httpClient, apiKey, time.Sleep)
	s.callGap = firstNonZeroDuration(tuning.CallGap, defaultCallGap)
	s.retryBackoffBase = firstNonZeroDuration(tuning.RetryBackoffBase, defaultRetryBackoffBase)
	s.retryBackoffMax = firstNonZeroDuration(tuning.RetryBackoffMax, defaultRetryBackoffMax)
	return s
}

func firstNonZeroDuration(v, fallback time.Duration) time.Duration {
	if v == 0 {
		return fallback
	}
	return v
}

func newSpeechSynthesizerForTest(httpClient *http.Client, apiKey string, backoffSleepFn func(time.Duration)) *SpeechSynthesizer {
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	return &SpeechSynthesizer{
		client:           httpClient,
		apiKey:           apiKey,
		backoffSleepFn:   backoffSleepFn,
		nowFn:            time.Now,
		callGap:          defaultCallGap,
		retryBackoffBase: defaultRetryBackoffBase,
		retryBackoffMax:  defaultRetryBackoffMax,
	}
}

// SynthesizeAll は texts を順に朗読音声へ変換し、セグメント単位の WAV 列（結合しない）を返す。
// retry 予算・callGap・RPD quota は Adapter 定数 = vendor 固有制約であり、
// 「1 episode 分の TTS 呼び出し群」を束ねて管理するのは Adapter の責務。
//
// @require texts の各要素は trim 後に非空。朗読本文のみ。
// @ensure 成功時は len(texts) と同数の非空・最小尺 WAV を返す（結合しない）。
// @ensure 呼び出し全体で Gemini 呼び出し合計を SynthesizeBudget 回以内へ抑える。
//
//	1 セグメントは min(MaxAttempts, 残予算) 回まで。合計が SynthesizeBudget へ達したら以降のセグメントは即 error。
func (s *SpeechSynthesizer) SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("synthesize", fmt.Errorf("client is nil"))
	}

	audios := make([]models.SpeechAudio, 0, len(texts))
	callsSpent := 0
	for i, text := range texts {
		remaining := SynthesizeBudget - callsSpent
		if remaining <= 0 {
			// why: ここへ来る前に必ず 1 セグメント以上を消費している。合計予算が尽きた。
			return nil, infraErr("synthesize_budget", fmt.Errorf(
				"gemini call budget exhausted at segment %d/%d: spent %d of %d", i+1, len(texts), callsSpent, SynthesizeBudget))
		}
		maxAttempts := MaxAttempts
		if remaining < maxAttempts {
			maxAttempts = remaining
		}
		audio, used, err := s.synthesizeOne(ctx, text, maxAttempts)
		callsSpent += used
		if err != nil {
			return nil, err
		}
		audios = append(audios, audio)
	}
	return audios, nil
}

// synthesizeOne は 1 本の text を最大 maxAttempts 回まで Gemini 呼び出しして朗読音声へ変換する。
// 戻りの int は実際に消費した Gemini 呼び出し回数（SynthesizeAll の残予算計算に使う）。
// callGap の lastCallAt 機構はそのまま流用するのでセグメントを跨いで効く。
func (s *SpeechSynthesizer) synthesizeOne(ctx context.Context, text string, maxAttempts int) (models.SpeechAudio, int, error) {
	backoffSleepFn := s.backoffSleepFn
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	nowFn := s.nowFn
	if nowFn == nil {
		nowFn = time.Now
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return models.SpeechAudio{}, 0, infraErr("validate_text", fmt.Errorf("text is empty after trim"))
	}
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	consecutiveSameOp := 0
	calls := 0
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		s.waitCallGap(backoffSleepFn, nowFn)
		pcm, retryable, suggestedWait, err := s.fetchPCM(ctx, trimmed)
		s.lastCallAt = nowFn()
		calls++
		if err == nil {
			wav, err := pcmToWAV(pcm)
			if err != nil {
				return models.SpeechAudio{}, calls, infraErr("pcm_to_wav", err)
			}
			return models.SpeechAudio{Content: wav}, calls, nil
		}
		// why: 同種 error（同じ *gemini.Error.Op）が retryable のまま 2 回連続したら、
		//      その本文に対しては決定論的に失敗しているとみなし打ち切る（Decision 2026-09-02T13-56-00）。
		//      Op が変われば連続数はリセットする。
		if retryable && sameGeminiOp(lastErr, err) {
			consecutiveSameOp++
		} else {
			consecutiveSameOp = 1
		}
		lastErr = err
		if !retryable || attempt == maxAttempts || consecutiveSameOp >= 2 {
			return models.SpeechAudio{}, calls, lastErr
		}
		wait := s.retryDelay(attempt)
		if suggestedWait > wait {
			wait = suggestedWait
		}
		backoffSleepFn(wait)
	}
	return models.SpeechAudio{}, calls, lastErr
}

// sameGeminiOp は 2 つの error がともに *gemini.Error で Op 文字列が一致するかを返す。
// 片方でも *gemini.Error でなければ false。
func sameGeminiOp(prev, cur error) bool {
	if prev == nil || cur == nil {
		return false
	}
	var prevErr, curErr *Error
	if !errors.As(prev, &prevErr) || !errors.As(cur, &curErr) {
		return false
	}
	return prevErr.Op == curErr.Op
}

func (s *SpeechSynthesizer) waitCallGap(sleepFn func(time.Duration), nowFn func() time.Time) {
	if s.lastCallAt.IsZero() {
		return
	}
	gap := s.effectiveCallGap()
	elapsed := nowFn().Sub(s.lastCallAt)
	if elapsed >= gap {
		return
	}
	sleepFn(gap - elapsed)
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

// parseRetryAfter は Retry-After 秒数 header を読む。無ければ 0。
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

func (s *SpeechSynthesizer) fetchPCM(ctx context.Context, transcript string) ([]byte, bool, time.Duration, error) {
	body, err := json.Marshal(interactionRequest{
		Model:  ModelID,
		Input:  buildInput(transcript),
		Format: responseFormat{Type: "audio"},
		GenerationConfig: generationConfig{
			SpeechConfig: []speechConfig{{Voice: VoiceName}},
		},
	})
	if err != nil {
		return nil, false, 0, infraErr("marshal_request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, EndpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, false, 0, infraErr("build_request", err)
	}
	req.Header.Set(geminiAPIKeyHeader, s.apiKey)
	res, err := s.client.Do(req)
	if err != nil {
		return nil, true, 0, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, true, 0, infraErr("read_body", err)
	}

	if prohibited := detectProhibitedContent(raw); prohibited {
		return nil, false, 0, infraErr("prohibited_content", fmt.Errorf("PROHIBITED_CONTENT"))
	}

	retryAfter := s.parseRetryAfter(res.Header)

	// why: MaxAttempts と公式 troubleshooting（429/503/5xx retry、400/403 は retry しない）に従い retryable を分岐する。
	switch {
	case res.StatusCode == http.StatusBadRequest, res.StatusCode == http.StatusForbidden:
		return nil, false, 0, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	case res.StatusCode == http.StatusTooManyRequests, res.StatusCode == http.StatusServiceUnavailable:
		return nil, true, retryAfter, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	case res.StatusCode >= 500:
		return nil, true, retryAfter, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	case res.StatusCode != http.StatusOK:
		return nil, false, 0, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}

	pcm, err := decodePCM(raw)
	if err != nil {
		// why: 公式 Limitation。audio 欠落 500 相当・minPCMBytes 未満の極小 PCM は
		//      いずれも一過性劣化として retry する。
		//      System で MaxAttempts 尽きたとき原因（finish_reason / safety / body 内 quota）を
		//      読めるよう、応答本文の bounded snippet を error に載せる。
		return nil, true, 0, infraErr("decode_pcm", fmt.Errorf("%w; response body: %s", err, bodySnippet(raw)))
	}
	return pcm, false, 0, nil
}

func buildInput(transcript string) string {
	return EnvelopePreamble + TranscriptLabel + transcript
}

func detectProhibitedContent(body []byte) bool {
	return strings.Contains(string(body), "PROHIBITED_CONTENT")
}

func decodePCM(body []byte) ([]byte, error) {
	var parsed interactionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	// why: 実 Interactions API（response_format audio）の audio base64 は
	//      steps[].content[].data に入る（run 33609034783 の診断本文で確定）。
	//      steps[0].content[0] 決め打ちにせず、空を飛ばして最初に見つかった data を採る。
	var data string
	for _, step := range parsed.Steps {
		for _, content := range step.Content {
			if trimmed := strings.TrimSpace(content.Data); trimmed != "" {
				data = trimmed
				break
			}
		}
		if data != "" {
			break
		}
	}
	if data == "" {
		// why: interactionResponse struct 経由の parse では audio 欠落の原因が読めない。
		//      body のトップレベルキー一覧を添え、レスポンス構造の想定違いを切り分ける。
		return nil, fmt.Errorf("output audio is missing%s", topLevelKeysHint(body))
	}
	pcm, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	if len(pcm) == 0 {
		return nil, fmt.Errorf("output audio is empty")
	}
	// why: Gemini が HTTP 200 で返す極小 PCM（len(pcm)==2 等）は実質無音の一過性劣化。
	//      audio 欠落 500 相当と同じ扱いで retry する（fetchPCM が decode_pcm op で retryable=true にする）。
	if len(pcm) < minPCMBytes {
		return nil, fmt.Errorf("output audio is too short: %d bytes < %d (%.1fs)", len(pcm), minPCMBytes, minSpeechDurationSec)
	}
	return pcm, nil
}

type interactionRequest struct {
	Model            string           `json:"model"`
	Input            string           `json:"input"`
	Format           responseFormat   `json:"response_format"`
	GenerationConfig generationConfig `json:"generation_config"`
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
