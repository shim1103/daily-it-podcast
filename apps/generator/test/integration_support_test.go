// Scope: Integration test 共通 support（Narrow / Broad 中立）
// 実物境界: なし（test double 組み立て helper のみ）
// Double: httptest TLS redirect・wire JSON fixture。
// @invariant dummy secret 実値は helper が error message へ出さない。
package test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
)

func minimalIntegrationGeminiPCM() []byte {
	// why: Adapter の最小尺閾値（0.5s）を超える長さ。これ未満だと極小 PCM として retry される。
	const sampleCount = 24000 // 1.0s 相当
	return make([]byte, sampleCount*2)
}

func writeIntegrationGeminiAudioResponse(t *testing.T, w http.ResponseWriter, pcm []byte) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, err := json.Marshal(map[string]any{
		"status": "completed",
		"steps": []map[string]any{
			{"content": []map[string]any{
				{"data": base64.StdEncoding.EncodeToString(pcm)},
			}},
		},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
