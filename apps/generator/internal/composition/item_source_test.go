package composition

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

func TestSourceMaxItemsForTopicCount_returnsZero_whenDraftTarget(t *testing.T) {
	t.Parallel()

	// Given/When: 本番既定の topicCount
	got := sourceMaxItemsForTopicCount(constants.DraftTopicCountTarget)

	// Then: Adapter 既定へ委譲するため 0
	if got != 0 {
		t.Fatalf("sourceMaxItemsForTopicCount(%d) = %d, want 0", constants.DraftTopicCountTarget, got)
	}
}

func TestSourceMaxItemsForTopicCount_scalesWithTopicCount_whenNotDraftTarget(t *testing.T) {
	t.Parallel()

	// Given: system-test 向けに絞った topicCount
	topicCount := 1

	// When
	got := sourceMaxItemsForTopicCount(topicCount)

	// Then: topic あたり件数 × topicCount
	want := topicCount * constants.SourceItemsPerTopic
	if got != want {
		t.Fatalf("sourceMaxItemsForTopicCount(%d) = %d, want %d", topicCount, got, want)
	}
}
