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
)

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
