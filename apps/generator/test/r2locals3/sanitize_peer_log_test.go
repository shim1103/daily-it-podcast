package r2locals3

import (
	"strings"
	"testing"
)

func TestSanitizePeerLog_redactsLocalCredentialsAndTruncates(t *testing.T) {
	t.Parallel()

	// Given: local peer 実値と長い log
	raw := strings.Repeat("x", 500) +
		" key=" + localAccessKeyID +
		" secret=" + localSecretAccessKey +
		" bucket=" + localBucket

	// When: sanitize する
	got := sanitizePeerLog(raw)

	// Then: 実値は残らず、長さ上限以下
	for _, leak := range []string{localAccessKeyID, localSecretAccessKey, localBucket} {
		if strings.Contains(got, leak) {
			t.Fatalf("sanitized log が %q を含む", leak)
		}
	}
	if len([]rune(got)) > peerLogMaxRunes {
		t.Fatalf("rune len=%d, want <= %d", len([]rune(got)), peerLogMaxRunes)
	}
	if !strings.Contains(got, "[redacted]") {
		t.Fatal("redacted marker が無い")
	}
}
