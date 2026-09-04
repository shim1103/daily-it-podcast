package gemini

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *SpeechSynthesizer) waitCallGap(sleepFn func(time.Duration), nowFn func() time.Time) {
	if s.lastCallAt.IsZero() {
		return
	}
	gap := s.effectiveCallGap()
	elapsed := nowFn().Sub(s.lastCallAt)
	if elapsed >= gap {
		return
	}
	sleepFn(gap - elapsed)
}

func (s *SpeechSynthesizer) effectiveCallGap() time.Duration {
	if s.callGap != 0 {
		return s.callGap
	}
	return defaultCallGap
}

func (s *SpeechSynthesizer) effectiveRetryBackoffBase() time.Duration {
	if s.retryBackoffBase != 0 {
		return s.retryBackoffBase
	}
	return defaultRetryBackoffBase
}

func (s *SpeechSynthesizer) effectiveRetryBackoffMax() time.Duration {
	if s.retryBackoffMax != 0 {
		return s.retryBackoffMax
	}
	return defaultRetryBackoffMax
}

func (s *SpeechSynthesizer) retryDelay(attempt int) time.Duration {
	// why: 公式は exponential。System の 429 対策で base を 60s・上限 3m にする。
	if attempt < 1 {
		attempt = 1
	}
	base := s.effectiveRetryBackoffBase()
	max := s.effectiveRetryBackoffMax()
	d := base << (attempt - 1)
	if d > max || d <= 0 {
		return max
	}
	return d
}

// parseRetryAfter は Retry-After 秒数 header を読む。無ければ 0。
func (s *SpeechSynthesizer) parseRetryAfter(h http.Header) time.Duration {
	raw := strings.TrimSpace(h.Get("Retry-After"))
	if raw == "" {
		return 0
	}
	sec, err := strconv.Atoi(raw)
	if err != nil || sec <= 0 {
		return 0
	}
	d := time.Duration(sec) * time.Second
	if max := s.effectiveRetryBackoffMax(); d > max {
		return max
	}
	return d
}
