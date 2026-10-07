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
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpdiag"
)

const (
	errorSource        = "gemini"
	geminiAPIKeyHeader = "x-goog-api-key"
)

// what: 公式 API errors ページが定める error.code（snake_case の機械可読 code）。
const (
	errorCodeQuotaExceeded     = "quota_exceeded"
	errorCodeRateLimitExceeded = "rate_limit_exceeded"
)

const prohibitedContentMarker = "PROHIBITED_CONTENT"

func infraErr(op string, err error) error {
	return adaptererror.New(errorSource, op, err)
}

type pcmFetchRetryKind int

const (
	// what: 再試行も fallback もせず、error をそのまま返す。
	pcmRetryNone pcmFetchRetryKind = iota
	// what: 待って再試行する。同種 error の連続は synthesizeOne が打ち切る。
	pcmRetryTransient
	// what: 回復が明示された 429。待って再試行する。
	pcmRetryRateLimited
	// what: 再試行せず fallback へ渡す。
	pcmRetryFallback
)

func (k pcmFetchRetryKind) retryable() bool {
	return k == pcmRetryTransient || k == pcmRetryRateLimited
}

// why: 再試行の使い切り（Transient・RateLimited）も Fallback と同じく fallback へ渡す。
func terminalError(kind pcmFetchRetryKind, err error) error {
	if kind == pcmRetryNone {
		return err
	}
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
}

// why: 429 の回復は response の message でなく、公式 error.code と Retry-After で見る。quota_exceeded は 400 でも返るので status を問わず見る（Decision generator-api-failure-retry-or-fallback）。
func (s *SpeechSynthesizer) classifyFailedStatus(status int, header http.Header, raw []byte) (pcmFetchRetryKind, time.Duration) {
	wait := s.parseRetryAfter(header)
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return pcmRetryFallback, 0
	case (status == http.StatusBadRequest || status == http.StatusTooManyRequests) && quotaExceeded(raw):
		return pcmRetryFallback, 0
	case status == http.StatusTooManyRequests:
		if wait > 0 || rateLimitExceeded(raw) {
			return pcmRetryRateLimited, wait
		}
		return pcmRetryFallback, 0
	case status >= http.StatusInternalServerError:
		return pcmRetryTransient, wait
	default:
		return pcmRetryNone, 0
	}
}

func (s *SpeechSynthesizer) fetchPCM(ctx context.Context, transcript string) ([]byte, pcmFetchRetryKind, time.Duration, error) {
	req, err := s.newSynthesizeRequest(ctx, transcript)
	if err != nil {
		return nil, pcmRetryNone, 0, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, pcmRetryTransient, 0, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, pcmRetryTransient, 0, infraErr("read_body", err)
	}

	if detectProhibitedContent(raw) {
		return nil, pcmRetryNone, 0, infraErr("prohibited_content", errors.New(prohibitedContentMarker))
	}

	if res.StatusCode != http.StatusOK {
		kind, wait := s.classifyFailedStatus(res.StatusCode, res.Header, raw)
		return nil, kind, wait, infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	}

	pcm, err := decodePCM(raw)
	if err != nil {
		// why: audio 欠落・極小 PCM は公式 Limitation の一過性劣化なので Transient にする。原因（finish_reason / safety / body 内 quota）を後から読めるよう、応答本文の snippet を error に載せる。
		return nil, pcmRetryTransient, 0, infraErr("decode_pcm", fmt.Errorf("%w; response body: %s", err, httpdiag.BodySnippet(raw)))
	}
	return pcm, pcmRetryNone, 0, nil
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

type errorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
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

func buildInput(transcript string) string {
	return EnvelopePreamble + TranscriptLabel + transcript
}

func detectProhibitedContent(body []byte) bool {
	return strings.Contains(string(body), prohibitedContentMarker)
}

func decodePCM(body []byte) ([]byte, error) {
	var parsed interactionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	data := firstAudioData(parsed)
	if data == "" {
		// why: struct 経由の parse では audio 欠落の原因が読めない。body のトップレベルキー一覧を添え、応答構造の想定違いを切り分ける。
		return nil, fmt.Errorf("output audio is missing%s", topLevelKeysHint(body))
	}
	pcm, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	if len(pcm) == 0 {
		return nil, fmt.Errorf("output audio is empty")
	}
	// why: HTTP 200 で返る極小 PCM は実質無音の一過性劣化。audio 欠落と同じく fetchPCM が Transient にする。
	if len(pcm) < minPCMBytes {
		return nil, fmt.Errorf("output audio is too short: %d bytes < %d (%.1fs)", len(pcm), minPCMBytes, minSpeechDurationSec)
	}
	return pcm, nil
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
