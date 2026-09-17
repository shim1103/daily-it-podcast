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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/httpdiag"
)

// why: この package の Infrastructure Error の発生源名。adaptererror.New へ渡す。
const errorSource = "gemini"

func infraErr(op string, err error) error {
	return adaptererror.New(errorSource, op, err)
}

// pcmFetchRetryKind は fetchPCM 取得失敗の再試行方針。geminiapi.fetchRetryKind と同型。
type pcmFetchRetryKind int

const (
	// pcmRetryNone は再試行しない（400 系、恒久的失敗を含む）。
	pcmRetryNone pcmFetchRetryKind = iota
	// pcmRetryTransient は Do error / 503 / 5xx / decode 失敗。synthesizeOne 側の
	// consecutiveSameOp 判定で打ち切りを制御する（fetchPCM 自身は回数を数えない）。
	pcmRetryTransient
	// pcmRetryRateLimited は 429。retry 打ち切り時に source 枯渇へ分類する。
	pcmRetryRateLimited
)

// fetchPCM は 1 回の Interactions API 呼び出しを実行し、(PCM, 再試行方針, 追加待ち, error) を返す。
//
// @ensure 401 / 403 相当、または response body が quota_exceeded 相当を示すときは
//
//	fmt.Errorf("%w: %w", port.ErrSourceExhausted, ...) で wrap した error を pcmRetryNone とともに返す。
//	この分類は fetch 時点で確定し、事後の文字列走査（旧 wrapIfSourceExhausted）には委ねない。
func (s *SpeechSynthesizer) fetchPCM(ctx context.Context, transcript string) ([]byte, pcmFetchRetryKind, time.Duration, error) {
	body, err := json.Marshal(interactionRequest{
		Model:  ModelID,
		Input:  buildInput(transcript),
		Format: responseFormat{Type: "audio"},
		GenerationConfig: generationConfig{
			SpeechConfig: []speechConfig{{Voice: VoiceName}},
		},
	})
	if err != nil {
		return nil, pcmRetryNone, 0, infraErr("marshal_request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, EndpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, pcmRetryNone, 0, infraErr("build_request", err)
	}
	req.Header.Set(geminiAPIKeyHeader, s.apiKey)
	res, err := s.client.Do(req)
	if err != nil {
		return nil, pcmRetryTransient, 0, infraErr("do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, pcmRetryTransient, 0, infraErr("read_body", err)
	}

	if prohibited := detectProhibitedContent(raw); prohibited {
		return nil, pcmRetryNone, 0, infraErr("prohibited_content", fmt.Errorf("PROHIBITED_CONTENT"))
	}

	retryAfter := s.parseRetryAfter(res.Header)

	// why: MaxAttempts と公式 troubleshooting（429/503/5xx retry、400/403 は retry しない）に従い retryable を分岐する。
	switch {
	case res.StatusCode == http.StatusUnauthorized, res.StatusCode == http.StatusForbidden:
		return nil, pcmRetryNone, 0, fmt.Errorf("%w: %w", port.ErrSourceExhausted,
			infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw))))
	case res.StatusCode == http.StatusBadRequest:
		if quotaExceeded(raw) {
			return nil, pcmRetryNone, 0, fmt.Errorf("%w: %w", port.ErrSourceExhausted,
				infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw))))
		}
		return nil, pcmRetryNone, 0, infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	case res.StatusCode == http.StatusTooManyRequests:
		return nil, pcmRetryRateLimited, retryAfter, infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	case res.StatusCode == http.StatusServiceUnavailable:
		return nil, pcmRetryTransient, retryAfter, infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	case res.StatusCode >= 500:
		return nil, pcmRetryTransient, retryAfter, infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	case res.StatusCode != http.StatusOK:
		return nil, pcmRetryNone, 0, infraErr("http_status", fmt.Errorf("status %d; response body: %s", res.StatusCode, httpdiag.BodySnippet(raw)))
	}

	pcm, err := decodePCM(raw)
	if err != nil {
		// why: 公式 Limitation。audio 欠落 500 相当・minPCMBytes 未満の極小 PCM は
		//      いずれも一過性劣化として retry する。
		//      System で MaxAttempts 尽きたとき原因（finish_reason / safety / body 内 quota）を
		//      読めるよう、応答本文の bounded snippet を error に載せる。
		return nil, pcmRetryTransient, 0, infraErr("decode_pcm", fmt.Errorf("%w; response body: %s", err, httpdiag.BodySnippet(raw)))
	}
	return pcm, pcmRetryNone, 0, nil
}

// errorResponse は公式 error response 形式（{"error": {"code": ..., "message": ...}}）の code だけを見る。
type errorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

// quotaExceeded は response body の error.code が quota_exceeded かを判定する。
// why: 公式 API errors ページが定義する error response 形式（error.code は snake_case の
//
//	machine-readable code）に従い JSON を parse する。
func quotaExceeded(raw []byte) bool {
	var parsed errorResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return false
	}
	return parsed.Error.Code == "quota_exceeded"
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
