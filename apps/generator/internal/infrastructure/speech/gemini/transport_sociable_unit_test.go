package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestBuildInput_wrapsTranscriptWithEnvelope_whenCallingProxy(t *testing.T) {

	// Given: 成功応答を返す Client Stub
	const transcript = "朗読する本文だけ"
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, audioInteractionResponse(minimalPCM())),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), transcript)

	// Then: envelope + Transcript ラベル + 本文が input へ入る
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d", len(rt.calls))
	}
	var req map[string]any
	if err := json.Unmarshal(rt.calls[0].Body, &req); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	input, _ := req["input"].(string)
	if !strings.Contains(input, EnvelopePreamble) {
		t.Fatalf("input missing preamble: %q", input)
	}
	if !strings.Contains(input, TranscriptLabel) {
		t.Fatalf("input missing transcript label: %q", input)
	}
	if !strings.Contains(input, transcript) {
		t.Fatalf("input missing transcript: %q", input)
	}
	genCfg, _ := req["generation_config"].(map[string]any)
	speechCfg, _ := genCfg["speech_config"].([]any)
	if len(speechCfg) != 1 {
		t.Fatalf("speech_config = %#v", speechCfg)
	}
	voiceCfg, _ := speechCfg[0].(map[string]any)
	if voiceCfg["voice"] != VoiceName {
		t.Fatalf("voice = %v, want %q", voiceCfg["voice"], VoiceName)
	}
	if req["model"] != ModelID {
		t.Fatalf("model = %v, want %q", req["model"], ModelID)
	}
}

func TestDecodePCM_extractsAudioFromStepsContentData_whenRealResponseShape(t *testing.T) {

	// Given: 実 Interactions API そっくりの応答（steps[].content[].data に audio base64）
	pcm := minimalPCM()
	body := map[string]any{
		"id":           "v1_real_shape",
		"object":       "interaction",
		"model":        ModelID,
		"status":       "completed",
		"service_tier": "standard",
		"created":      "2026-09-02T08:36:33Z",
		"updated":      "2026-09-02T08:36:33Z",
		"usage": map[string]any{
			"total_tokens":        125,
			"total_output_tokens": 99,
			"output_tokens_by_modality": []map[string]any{
				{"modality": "audio", "tokens": 99},
			},
		},
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(pcm)},
			}},
		},
	}
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body:   jsonBody(t, body),
	})

	// When: Synthesize する
	got, err := synth.synthTestOne(context.Background(), "実応答形テスト")

	// Then: steps 構造から audio を取り出し非空 WAV を返す
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if !isWAV(got.Content) {
		t.Fatalf("Content is not WAV: %d bytes", len(got.Content))
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}

func TestDecodePCM_returnsInfrastructureError_whenOutputAudioMissingOnOK(t *testing.T) {

	// Given: HTTP 200 だが output_audio が無い（同種 retryable = Op "decode_pcm"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body:   jsonBody(t, map[string]any{"status": "ok"}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "audio 欠落")

	// Then: 同種 2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
	// Then: error 文言に body のトップレベルキー一覧が載り、fixture の "status" が含まれる
	if !strings.Contains(err.Error(), "top-level keys:") {
		t.Fatalf("Error() = %q, want it to contain %q", err.Error(), "top-level keys:")
	}
	if !strings.Contains(err.Error(), "status") {
		t.Fatalf("Error() = %q, want it to list top-level key %q", err.Error(), "status")
	}
}

func TestDecodePCM_retriesTooShortPCM_asTransientDecodeFailure(t *testing.T) {

	// Given: HTTP 200 だが尺が minPCMBytes 未満の極小 PCM（1 サンプル）。
	//        Gemini の一過性劣化なので decode_pcm 相当の retryable として retry される。
	tinyPCM := []byte{0x00, 0x00}
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"steps": []map[string]any{
					{"content": []map[string]any{
						{"data": base64.StdEncoding.EncodeToString(tinyPCM)},
					}},
				},
			}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "極小 PCM")

	// Then: 同種 retryable（Op "decode_pcm"）2 連続で打ち切り Infrastructure Error
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *gemini.Error", err, err)
	}
	if infra.Op != "decode_pcm" {
		t.Fatalf("Op = %q, want %q（極小 PCM は decode_pcm 系 retryable）", infra.Op, "decode_pcm")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestDecodePCM_returnsInfrastructureError_whenResponseBodyInvalidJSON(t *testing.T) {

	// Given: 壊れた JSON（同種 retryable = Op "decode_pcm"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{status: http.StatusOK, body: []byte(`not-json`)}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "decode 失敗")

	// Then: 同種 2 連続で打ち切り error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestDecodePCM_returnsInfrastructureError_whenBase64Invalid(t *testing.T) {

	// Given: 不正 base64（同種 retryable = Op "decode_pcm"）
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"steps": []map[string]any{
					{"content": []map[string]any{
						{"data": "!!!not-base64!!!"},
					}},
				},
			}),
		}
	}
	synth, rt := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "bad b64")

	// Then: 同種 2 連続で打ち切り error
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("call count = %d, want 2（同種 2 連続打ち切り）", len(rt.calls))
	}
}

func TestDecodePCM_returnsInfrastructureError_whenPCMLengthOdd(t *testing.T) {

	// Given: 最小尺は満たすが 16-bit 非整列（奇数 byte）の PCM。
	//        極小 PCM の retryable 判定は抜け、pcmToWAV の非整列拒否（非 retry）に落ちる。
	oddPCM := make([]byte, minPCMBytes+1)
	synth, rt := newFakeSynthesizer(fakeClientResponse{
		status: http.StatusOK,
		body: jsonBody(t, map[string]any{
			"steps": []map[string]any{
				{"content": []map[string]any{
					{"data": base64.StdEncoding.EncodeToString(oddPCM)},
				}},
			},
		}),
	})

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "奇数 pcm")

	// Then: pcm 変換失敗（非 retry）
	if err == nil {
		t.Fatal("expected error")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(rt.calls))
	}
}
