package composition

import (
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewEpisodeID_returnsRFC4122V4String(t *testing.T) {
	t.Parallel()

	// Given: newEpisodeID（crypto/rand UUIDv4 実体）

	// When: 2 回発行する
	a := newEpisodeID()
	b := newEpisodeID()

	// Then: 2 つとも RFC4122 v4 の書式（version nibble = 4、variant nibble ∈ 8..b）
	if !uuidV4Pattern.MatchString(a) {
		t.Fatalf("newEpisodeID() = %q, want RFC4122 v4", a)
	}
	if !uuidV4Pattern.MatchString(b) {
		t.Fatalf("newEpisodeID() = %q, want RFC4122 v4", b)
	}

	// Then: 毎回異なる（乱数由来）
	if a == b {
		t.Fatalf("newEpisodeID() は毎回同じ値を返した: %q", a)
	}
}
