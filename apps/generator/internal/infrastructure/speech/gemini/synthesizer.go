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

type SpeechSynthesizer struct {
	client         *http.Client
	apiKey         string
	backoffSleepFn func(time.Duration) // why: test の並列実行と共存するため package global に置かない
	lastCallAt     time.Time
	nowFn          func() time.Time
}

// NewSpeechSynthesizer は Gemini TTS Adapter を返す。
//
// @require httpClient != nil
// @ensure apiKey は x-goog-api-key header にだけ使い、保存元の知識は持たない。
func NewSpeechSynthesizer(httpClient *http.Client, apiKey string) *SpeechSynthesizer {
	return newSpeechSynthesizerForTest(httpClient, apiKey, time.Sleep)
}

func newSpeechSynthesizerForTest(httpClient *http.Client, apiKey string, backoffSleepFn func(time.Duration)) *SpeechSynthesizer {
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	return &SpeechSynthesizer{
		client:         httpClient,
		apiKey:         apiKey,
		backoffSleepFn: backoffSleepFn,
		nowFn:          time.Now,
	}
}

func (s *SpeechSynthesizer) Synthesize(ctx context.Context, text string) (models.SpeechAudio, error) {
	if s == nil || s.client == nil {
		return models.SpeechAudio{}, infraErr("synthesize", fmt.Errorf("client is nil"))
	}
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
		return models.SpeechAudio{}, infraErr("validate_text", fmt.Errorf("text is empty after trim"))
	}

	var lastErr error
	consecutiveSameOp := 0
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		s.waitCallGap(backoffSleepFn, nowFn)
		pcm, retryable, suggestedWait, err := s.fetchPCM(ctx, trimmed)
		s.lastCallAt = nowFn()
		if err == nil {
			wav, err := pcmToWAV(pcm)
			if err != nil {
				return models.SpeechAudio{}, infraErr("pcm_to_wav", err)
			}
			return models.SpeechAudio{Content: wav}, nil
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
		if !retryable || attempt == MaxAttempts || consecutiveSameOp >= 2 {
			return models.SpeechAudio{}, lastErr
		}
		wait := retryDelay(attempt)
		if suggestedWait > wait {
			wait = suggestedWait
		}
		backoffSleepFn(wait)
	}
	return models.SpeechAudio{}, lastErr
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
	elapsed := nowFn().Sub(s.lastCallAt)
	if elapsed >= callGap {
		return
	}
	sleepFn(callGap - elapsed)
}

func retryDelay(attempt int) time.Duration {
	// why: 公式は exponential。System の 429 対策で base を 60s・上限 3m にする。
	if attempt < 1 {
		attempt = 1
	}
	d := retryBackoffBase << (attempt - 1)
	if d > retryBackoffMax || d <= 0 {
		return retryBackoffMax
	}
	return d
}

// parseRetryAfter は Retry-After 秒数 header を読む。無ければ 0。
func parseRetryAfter(h http.Header) time.Duration {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	if raw == "" {
		return 0
	}
	sec, err := strconv.Atoi(raw)
	if err != nil || sec <= 0 {
		return 0
	}
	d := time.Duration(sec) * time.Second
	if d > retryBackoffMax {
		return retryBackoffMax
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

	retryAfter := parseRetryAfter(res.Header)

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
		// why: 公式 Limitation。audio 欠落 500 相当は一過性として retry する。
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
