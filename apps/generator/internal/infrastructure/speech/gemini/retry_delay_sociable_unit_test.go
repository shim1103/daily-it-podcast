package gemini

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryDelay_growsFromBaseAndCapsAtMax(t *testing.T) {
	t.Parallel()

	if got := retryDelay(1); got != retryBackoffBase {
		t.Fatalf("retryDelay(1) = %v, want %v", got, retryBackoffBase)
	}
	if got := retryDelay(2); got != 2*retryBackoffBase {
		t.Fatalf("retryDelay(2) = %v, want %v", got, 2*retryBackoffBase)
	}
	if got := retryDelay(10); got != retryBackoffMax {
		t.Fatalf("retryDelay(10) = %v, want %v", got, retryBackoffMax)
	}
	if retryBackoffBase < 60*time.Second {
		t.Fatalf("retryBackoffBase = %v, want >= 60s（429 対策）", retryBackoffBase)
	}
}

func TestParseRetryAfter_returnsDuration_whenSecondsHeaderPresent(t *testing.T) {
	t.Parallel()

	h := http.Header{}
	h.Set("Retry-After", "90")
	if got := parseRetryAfter(h); got != 90*time.Second {
		t.Fatalf("parseRetryAfter = %v, want 90s", got)
	}
}

func TestParseRetryAfter_returnsZero_whenHeaderMissingOrInvalid(t *testing.T) {
	t.Parallel()

	if got := parseRetryAfter(http.Header{}); got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
	h := http.Header{}
	h.Set("Retry-After", "nope")
	if got := parseRetryAfter(h); got != 0 {
		t.Fatalf("invalid = %v, want 0", got)
	}
}
