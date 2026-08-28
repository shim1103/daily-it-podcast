package composition

import (
	"testing"
)

func TestCursorCommandInheritedEnvNameAllow_returnsCopy_withoutMutatingSSoT(t *testing.T) {
	t.Parallel()

	// Given: 公開 allowlist
	first := CursorCommandInheritedEnvNameAllow()
	if len(first) == 0 {
		t.Fatal("CursorCommandInheritedEnvNameAllow() = empty, want non-empty")
	}
	original := first[0]

	// When: 呼び出し側が戻りを書き換える
	first[0] = "MUTATED"

	// Then: SSoT は壊れない
	second := CursorCommandInheritedEnvNameAllow()
	if second[0] != original {
		t.Fatalf("SSoT mutated: second[0] = %q, want %q", second[0], original)
	}
}
