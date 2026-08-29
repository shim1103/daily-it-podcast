package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	return &SpeechSynthesizer{client: httpClient, apiKey: apiKey, backoffSleepFn: backoffSleepFn}
}

func (s *SpeechSynthesizer) Synthesize(ctx context.Context, text string) (models.SpeechAudio, error) {
	if s == nil || s.client == nil {
		return models.SpeechAudio{}, infraErr("synthesize", fmt.Errorf("client is nil"))
	}
	backoffSleepFn := s.backoffSleepFn
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return models.SpeechAudio{}, infraErr("validate_text", fmt.Errorf("text is empty after trim"))
	}

	var lastErr error
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		pcm, retryable, err := s.fetchPCM(ctx, trimmed)
		if err == nil {
			wav, err := pcmToWAV(pcm)
			if err != nil {
				return models.SpeechAudio{}, infraErr("pcm_to_wav", err)
			}
			return models.SpeechAudio{Content: wav}, nil
		}
		lastErr = err
		if !retryable || attempt == MaxAttempts {
			return models.SpeechAudio{}, lastErr
		}
		backoffSleepFn(retryDelay(attempt))
	}
	return models.SpeechAudio{}, lastErr
}

func retryDelay(attempt int) time.Duration {
	// why: 公式 troubleshooting の exponential backoff（1s, 2s, 4s…）に合わせる。
	if attempt < 1 {
		attempt = 1
	}
	return time.Second << (attempt - 1)
}

func (s *SpeechSynthesizer) fetchPCM(ctx context.Context, transcript string) ([]byte, bool, error) {
	body, err := json.Marshal(interactionRequest{
		Model:  ModelID,
		Input:  buildInput(transcript),
		Format: responseFormat{Type: "audio"},
		GenerationConfig: generationConfig{
			SpeechConfig: []speechConfig{{Voice: VoiceName}},
		},
	})
	if err != nil {
		return nil, false, infraErr("marshal_request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, EndpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, false, infraErr("build_request", err)
	}
	req.Header.Set(geminiAPIKeyHeader, s.apiKey)
	res, err := s.client.Do(req)
	if err != nil {
		return nil, true, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, true, infraErr("read_body", err)
	}

	if prohibited := detectProhibitedContent(raw); prohibited {
		return nil, false, infraErr("prohibited_content", fmt.Errorf("PROHIBITED_CONTENT"))
	}

	// why: MaxAttempts と公式 troubleshooting（429/503/5xx retry、400/403 は retry しない）に従い retryable を分岐する。
	switch {
	case res.StatusCode == http.StatusBadRequest, res.StatusCode == http.StatusForbidden:
		return nil, false, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	case res.StatusCode == http.StatusTooManyRequests, res.StatusCode == http.StatusServiceUnavailable:
		return nil, true, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	case res.StatusCode >= 500:
		return nil, true, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	case res.StatusCode != http.StatusOK:
		return nil, false, infraErr("http_status", fmt.Errorf("status %d", res.StatusCode))
	}

	pcm, err := decodePCM(raw)
	if err != nil {
		// why: 公式 Limitation。audio 欠落 500 相当は一過性として retry する。
		return nil, true, infraErr("decode_pcm", err)
	}
	return pcm, false, nil
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
	data := strings.TrimSpace(parsed.OutputAudio.Data)
	if data == "" {
		return nil, fmt.Errorf("output audio is missing")
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
	OutputAudio outputAudio `json:"output_audio"`
}

type outputAudio struct {
	Data string `json:"data"`
}
