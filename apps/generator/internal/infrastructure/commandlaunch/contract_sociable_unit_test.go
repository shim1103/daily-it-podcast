package commandlaunch

import (
	"testing"
)

func TestInheritedEnvNameAllow_returnsDefensiveCopy_whenCallerModifiesReturnValue(t *testing.T) {
	// Given: InheritedEnvNameAllow の戻り値
	first := InheritedEnvNameAllow()

	// When: 戻り値を変更する
	first[0] = "MUTATED"

	// Then: 次の呼び出しが原値を返す（独立したコピー）
	second := InheritedEnvNameAllow()
	if second[0] == "MUTATED" {
		t.Fatalf("InheritedEnvNameAllow()[0] = %q, want immutable value after modification", second[0])
	}
	if second[0] != "PATH" {
		t.Fatalf("InheritedEnvNameAllow()[0] = %q, want %q", second[0], "PATH")
	}
}
