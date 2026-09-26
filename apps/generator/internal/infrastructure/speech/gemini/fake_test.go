package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

func minimalPCM() []byte {
	// why: 24 kHz / 16-bit / mono。最小尺閾値（minPCMBytes = 0.5s）を超える長さにする。
	//      これ未満だと decodePCM が「極小 PCM」として retryable な失敗に落とす。
	const sampleCount = 24000 // 1.0s 相当（24000 * 2 = 48000 bytes > minPCMBytes 24000）
	pcm := make([]byte, sampleCount*2)
	return pcm
}

func audioInteractionResponse(pcm []byte) map[string]any {
	return map[string]any{
		"status": "completed",
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(pcm)},
			}},
		},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

func isWAV(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}

// fakeClientCall は fakeRoundTripper が観測した request 1 件分。
type fakeClientCall struct {
	Method string
	URL    string
	Body   []byte
}

// fakeRoundTripper は境界 I/O なしで http.RoundTripper を満たす Spy。
// 呼び出し順に responses を返し、各 request を記録する。
type fakeRoundTripper struct {
	responses []fakeClientResponse
	calls     []fakeClientCall
}

type fakeClientResponse struct {
	status int
	body   []byte
	err    error
}

func (rt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}
	rt.calls = append(rt.calls, fakeClientCall{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   body,
	})
	index := len(rt.calls) - 1
	if index >= len(rt.responses) {
		return nil, fmt.Errorf("fakeRoundTripper: no response configured for call %d", index)
	}
	resp := rt.responses[index]
	if resp.err != nil {
		return nil, resp.err
	}
	rec := httptest.NewRecorder()
	rec.WriteHeader(resp.status)
	if resp.body != nil {
		_, _ = rec.Write(resp.body)
	}
	return rec.Result(), nil
}

func jsonBody(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return raw
}

func newFakeSynthesizer(responses ...fakeClientResponse) (*SpeechSynthesizer, *fakeRoundTripper) {
	rt := &fakeRoundTripper{responses: responses}
	synth := newSpeechSynthesizerForTest(&http.Client{Transport: rt}, "gemini-fake-key", func(time.Duration) {})
	return synth, rt
}

// synthTestOne は 1 セグメント分の retry ループ（synthesizeOne）を MaxAttempts 上限で叩く test helper。
// 単一セグメントの loop 挙動を検証する既存 test 用。合計予算・入口ガードの検証は SynthesizeAll の test が持つ。
func (s *SpeechSynthesizer) synthTestOne(ctx context.Context, text string) (models.SpeechAudio, error) {
	audio, _, err := s.synthesizeOne(ctx, text, MaxAttempts)
	return audio, err
}
