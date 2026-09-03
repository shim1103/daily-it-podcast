package gemini

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// why: System で `decode_pcm: output audio is missing` が MaxAttempts 尽きても
//
//	応答本文が error に載らず原因（finish_reason / safety / body 内 quota）が読めない（run 33581258235）。
//	HTTP 200 で audio 欠落のとき、error は本文の bounded snippet を含む。
func TestSynthesize_includesResponseBodySnippet_whenOutputAudioMissingOnOK(t *testing.T) {

	// Given: HTTP 200 だが output_audio が無く、代わりに finish_reason を持つ本文
	const marker = "SAFETY_BLOCKLIST_TRIGGERED"
	responses := make([]fakeClientResponse, MaxAttempts)
	for i := range responses {
		responses[i] = fakeClientResponse{
			status: http.StatusOK,
			body: jsonBody(t, map[string]any{
				"finish_reason": marker,
				"note":          "no audio produced",
			}),
		}
	}
	synth, _ := newFakeSynthesizer(responses...)

	// When: Synthesize する
	_, err := synth.synthTestOne(context.Background(), "audio 欠落の原因を知りたい")

	// Then: error 文言に本文の marker が含まれる
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), marker) {
		t.Fatalf("Error() = %q, want it to contain response body marker %q", err.Error(), marker)
	}
}

func TestBodySnippet_truncatesLongBodyAndStripsNewlines(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", bodySnippetMax*2)
	got := bodySnippet([]byte(long))
	if len(got) > bodySnippetMax+len(bodySnippetEllipsis) {
		t.Fatalf("len(bodySnippet) = %d, want <= %d", len(got), bodySnippetMax+len(bodySnippetEllipsis))
	}
	if !strings.HasSuffix(got, bodySnippetEllipsis) {
		t.Fatalf("bodySnippet = %q, want ellipsis suffix", got)
	}

	multiline := "line1\nline2\r\nline3"
	if strings.ContainsAny(bodySnippet([]byte(multiline)), "\r\n") {
		t.Fatalf("bodySnippet kept newlines: %q", bodySnippet([]byte(multiline)))
	}
}
