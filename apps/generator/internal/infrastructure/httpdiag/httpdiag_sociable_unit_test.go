package httpdiag

import (
	"strings"
	"testing"
)

func TestBodySnippet_returnsTrimmedBodyUnchanged_whenShort(t *testing.T) {
	t.Parallel()

	// Given: maxRunes 以下で前後に空白のある本文
	raw := []byte("  {\"error\":\"UNAUTHENTICATED\"}  ")

	// When: snippet へ落とす
	got := BodySnippet(raw)

	// Then: trim だけされ ellipsis は付かない
	if got != `{"error":"UNAUTHENTICATED"}` {
		t.Fatalf("BodySnippet() = %q, want %q", got, `{"error":"UNAUTHENTICATED"}`)
	}
}

func TestBodySnippet_truncatesWithEllipsis_whenBodyExceedsMax(t *testing.T) {
	t.Parallel()

	// Given: maxRunes を超える本文
	got := BodySnippet([]byte(strings.Repeat("a", maxRunes*2)))

	// Then: maxRunes + ellipsis 長を超えず ellipsis で終わる
	if len(got) != maxRunes+len(ellipsis) {
		t.Fatalf("len(BodySnippet()) = %d, want %d", len(got), maxRunes+len(ellipsis))
	}
	if !strings.HasSuffix(got, ellipsis) {
		t.Fatalf("BodySnippet() = %q, want ellipsis suffix", got)
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
