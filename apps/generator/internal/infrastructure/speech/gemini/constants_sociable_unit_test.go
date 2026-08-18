package gemini_test

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

func TestMaxAttempts_isAtLeastOne_whenConstantsDefined(t *testing.T) {
	// Given: Adapter 定数が定義されている
	// When: 省略（定数参照のみ）

	// Then: 無限 retry にならない
	if gemini.MaxAttempts < 1 {
		t.Fatalf("MaxAttempts = %d, want >= 1", gemini.MaxAttempts)
	}
}
