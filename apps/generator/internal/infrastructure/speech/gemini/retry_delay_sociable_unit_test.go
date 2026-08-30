package gemini

import (
	"testing"
	"time"
)

func TestRetryDelay_growsFromBaseAndCapsAtMax(t *testing.T) {
	t.Parallel()

	// Given / When / Then: 1 回目は base、以降は倍、上限で頭打ち
	if got := retryDelay(1); got != retryBackoffBase {
		t.Fatalf("retryDelay(1) = %v, want %v", got, retryBackoffBase)
	}
	if got := retryDelay(2); got != 2*retryBackoffBase {
		t.Fatalf("retryDelay(2) = %v, want %v", got, 2*retryBackoffBase)
	}
	if got := retryDelay(10); got != retryBackoffMax {
		t.Fatalf("retryDelay(10) = %v, want %v", got, retryBackoffMax)
	}
	if retryBackoffBase < 15*time.Second {
		t.Fatalf("retryBackoffBase = %v, want >= 15s（429 対策）", retryBackoffBase)
	}
}
