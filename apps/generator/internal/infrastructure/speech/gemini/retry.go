package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// why: 再試行できる失敗が同種のまま続くのは、その本文に対して決定論的に失敗しているとみなすため（Decision 2026-09-02T13-56-00）。
const maxConsecutiveSameOp = 2

// failureStreak は直近の失敗と、同じ *adaptererror.Error.Op が再試行可能のまま連続した回数を持つ。
type failureStreak struct {
	last  error
	count int
}

func (f *failureStreak) record(kind pcmFetchRetryKind, err error) {
	if kind.retryable() && sameGeminiOp(f.last, err) {
		f.count++
	} else {
		f.count = 1
	}
	f.last = err
}

func (f *failureStreak) repeatsSameOp() bool {
	return f.count >= maxConsecutiveSameOp
}

// what: 戻りの int は実際に消費した Gemini 呼び出し回数。SynthesizeAll が残予算の計算に使う。
func (s *SpeechSynthesizer) synthesizeOne(ctx context.Context, text string, maxAttempts int) (models.SpeechAudio, int, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return models.SpeechAudio{}, 0, infraErr("validate_text", fmt.Errorf("text is empty after trim"))
	}
	maxAttempts = max(maxAttempts, 1)
	sleep, now := s.sleepFunc(), s.nowFunc()

	var streak failureStreak
	calls := 0
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		s.waitCallGap(sleep, now)
		pcm, kind, suggestedWait, err := s.fetchPCM(ctx, trimmed)
		s.lastCallAt = now()
		calls++
		if err == nil {
			audio, err := toSpeechAudio(pcm)
			return audio, calls, err
		}
		streak.record(kind, err)
		if !kind.retryable() || attempt == maxAttempts || streak.repeatsSameOp() {
			return models.SpeechAudio{}, calls, terminalError(kind, err)
		}
		s.retry.Retry("synthesize_speech", attempt, maxAttempts, err.Error())
		sleep(max(s.retryDelay(attempt), suggestedWait))
	}
	return models.SpeechAudio{}, calls, streak.last
}

func toSpeechAudio(pcm []byte) (models.SpeechAudio, error) {
	wav, err := pcmToWAV(pcm)
	if err != nil {
		return models.SpeechAudio{}, infraErr("pcm_to_wav", err)
	}
	return models.SpeechAudio{Content: wav, DurationSec: pcmDurationSec(pcm)}, nil
}

func (s *SpeechSynthesizer) sleepFunc() func(time.Duration) {
	if s.backoffSleepFn == nil {
		return time.Sleep
	}
	return s.backoffSleepFn
}

func (s *SpeechSynthesizer) nowFunc() func() time.Time {
	if s.nowFn == nil {
		return time.Now
	}
	return s.nowFn
}
