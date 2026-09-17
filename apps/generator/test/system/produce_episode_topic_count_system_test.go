package system

import "testing"

func TestSystemTestTopicCount_usesExplicitPositiveValue(t *testing.T) {
	// Given
	t.Setenv(systemTestTopicCountEnv, "2")

	// When
	got := systemTestTopicCount()

	// Then
	if got != 2 {
		t.Fatalf("systemTestTopicCount() = %d、期待値 = 2", got)
	}
}
