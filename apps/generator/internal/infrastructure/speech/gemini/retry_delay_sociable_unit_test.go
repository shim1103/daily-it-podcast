package gemini

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryDelay_growsFromBaseAndCapsAtMax(t *testing.T) {
	t.Parallel()

	s := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key")

	if got := s.retryDelay(1); got != defaultRetryBackoffBase {
		t.Fatalf("retryDelay(1) = %v, want %v", got, defaultRetryBackoffBase)
	}
	if got := s.retryDelay(2); got != 2*defaultRetryBackoffBase {
		t.Fatalf("retryDelay(2) = %v, want %v", got, 2*defaultRetryBackoffBase)
	}
	if got := s.retryDelay(10); got != defaultRetryBackoffMax {
		t.Fatalf("retryDelay(10) = %v, want %v", got, defaultRetryBackoffMax)
	}
	if defaultRetryBackoffBase < 60*time.Second {
		t.Fatalf("defaultRetryBackoffBase = %v, want >= 60s（429 対策）", defaultRetryBackoffBase)
	}
}

func TestParseRetryAfter_returnsDuration_whenSecondsHeaderPresent(t *testing.T) {
	t.Parallel()

	s := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key")

	h := http.Header{}
	h.Set("Retry-After", "90")
	if got := s.parseRetryAfter(h); got != 90*time.Second {
		t.Fatalf("parseRetryAfter = %v, want 90s", got)
	}
}

func TestParseRetryAfter_returnsZero_whenHeaderMissingOrInvalid(t *testing.T) {
	t.Parallel()

	s := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key")

	if got := s.parseRetryAfter(http.Header{}); got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
	h := http.Header{}
	h.Set("Retry-After", "nope")
	if got := s.parseRetryAfter(h); got != 0 {
		t.Fatalf("invalid = %v, want 0", got)
	}
}
