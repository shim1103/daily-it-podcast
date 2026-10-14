package httpdiag

import (
	"strings"
	"testing"
)

func TestBodySnippet_returnsTrimmedBodyUnchanged_whenShort(t *testing.T) {
	t.Parallel()

	// Given: 前後に空白のある本文
	raw := []byte("  {\"error\":\"UNAUTHENTICATED\"}  ")

	// When: snippet へ落とす
	got := BodySnippet(raw)

	// Then: trim だけされる
	if got != `{"error":"UNAUTHENTICATED"}` {
		t.Fatalf("BodySnippet() = %q, want %q", got, `{"error":"UNAUTHENTICATED"}`)
	}
}

func TestBodySnippet_keepsFullBody_whenBodyIsLong(t *testing.T) {
	t.Parallel()

	// Given: 長い本文（旧 maxRunes=400 を超える長さ）
	body := strings.Repeat("a", 800)

	// When: snippet へ落とす
	got := BodySnippet([]byte(body))

	// Then: 切り詰めず全文を返す
	if got != body {
		t.Fatalf("BodySnippet() len = %d, want %d", len(got), len(body))
	}
}

func TestBodySnippet_collapsesNewlinesToSpace_whenBodyIsMultiline(t *testing.T) {
	t.Parallel()

	// Given: \n / \r\n / \r を含む本文
	got := BodySnippet([]byte("line1\nline2\r\nline3\rline4"))

	// Then: 改行は空白へ潰れ 1 行になる
	if strings.ContainsAny(got, "\r\n") {
		t.Fatalf("BodySnippet() kept newlines: %q", got)
	}
	if got != "line1 line2 line3 line4" {
		t.Fatalf("BodySnippet() = %q, want %q", got, "line1 line2 line3 line4")
	}
}
