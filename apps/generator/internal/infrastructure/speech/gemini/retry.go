package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// synthesizeOne は 1 本の text を最大 maxAttempts 回まで Gemini 呼び出しして朗読音声へ変換する。
// 戻りの int は実際に消費した Gemini 呼び出し回数（SynthesizeAll の残予算計算に使う）。
// callGap の lastCallAt 機構はそのまま流用するのでセグメントを跨いで効く。
func (s *SpeechSynthesizer) synthesizeOne(ctx context.Context, text string, maxAttempts int) (models.SpeechAudio, int, error) {
	backoffSleepFn := s.backoffSleepFn
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	nowFn := s.nowFn
	if nowFn == nil {
		nowFn = time.Now
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return models.SpeechAudio{}, 0, infraErr("validate_text", fmt.Errorf("text is empty after trim"))
	}
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	consecutiveSameOp := 0
	calls := 0
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		s.waitCallGap(backoffSleepFn, nowFn)
		pcm, retryable, suggestedWait, err := s.fetchPCM(ctx, trimmed)
		s.lastCallAt = nowFn()
		calls++
		if err == nil {
			wav, err := pcmToWAV(pcm)
			if err != nil {
				return models.SpeechAudio{}, calls, infraErr("pcm_to_wav", err)
			}
			return models.SpeechAudio{Content: wav}, calls, nil
		}
		// why: 同種 error（同じ *gemini.Error.Op）が retryable のまま 2 回連続したら、
		//      その本文に対しては決定論的に失敗しているとみなし打ち切る（Decision 2026-09-02T13-56-00）。
		//      Op が変われば連続数はリセットする。
		if retryable && sameGeminiOp(lastErr, err) {
			consecutiveSameOp++
		} else {
			consecutiveSameOp = 1
		}
		lastErr = err
		if !retryable || attempt == maxAttempts || consecutiveSameOp >= 2 {
			return models.SpeechAudio{}, calls, lastErr
		}
		wait := s.retryDelay(attempt)
		if suggestedWait > wait {
			wait = suggestedWait
		}
		backoffSleepFn(wait)
	}
	return models.SpeechAudio{}, calls, lastErr
}
