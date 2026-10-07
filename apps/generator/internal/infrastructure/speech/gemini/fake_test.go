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

type retryReporterSpy struct {
	calls int
}

func (s *retryReporterSpy) Retry(step string, attempt, maxAttempts int, reason string) {
	s.calls++
}

// why: minPCMBytes 未満だと decodePCM が極小 PCM として一過性の失敗に落とすため、閾値を超える長さにする。
func minimalPCM() []byte {
	return make([]byte, 2*minPCMBytes)
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

func isWAV(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}

type fakeClientCall struct {
	Method string
	URL    string
	Body   []byte
}

type fakeRoundTripper struct {
	responses []fakeClientResponse
	calls     []fakeClientCall
}

type fakeClientResponse struct {
	status int
	header http.Header
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
	for key, values := range resp.header {
		for _, v := range values {
			rec.Header().Add(key, v)
		}
	}
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
	synth := newSpeechSynthesizerForTest(&http.Client{Transport: rt}, "gemini-fake-key", func(time.Duration) {}, &retryReporterSpy{})
	return synth, rt
}

// why: 合計予算・入口ガードは SynthesizeAll の test が持つので、ここは 1 セグメントの retry loop だけを MaxAttempts 上限で叩く。
func (s *SpeechSynthesizer) synthTestOne(ctx context.Context, text string) (models.SpeechAudio, error) {
	audio, _, err := s.synthesizeOne(ctx, text, MaxAttempts)
	return audio, err
}
