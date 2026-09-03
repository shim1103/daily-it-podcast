package gemini

import (
	"net/http"
	"time"
)

// Tuning は SpeechSynthesizer の待機系パラメータの注入値。
// ゼロ値 field は既定値（default*）へフォールバックする。
type Tuning struct {
	CallGap          time.Duration
	RetryBackoffBase time.Duration
	RetryBackoffMax  time.Duration
}

// NewSpeechSynthesizer は Gemini TTS Adapter を返す。待機系パラメータは既定値。
//
// @require httpClient != nil
// @ensure apiKey は x-goog-api-key header にだけ使い、保存元の知識は持たない。
func NewSpeechSynthesizer(httpClient *http.Client, apiKey string) *SpeechSynthesizer {
	return NewSpeechSynthesizerWithTuning(httpClient, apiKey, Tuning{})
}

// NewSpeechSynthesizerWithTuning は待機系パラメータを注入できる constructor。
// rate 計測（system && ratemeasure）専用の差し替え口。
//
// @require httpClient != nil
// @ensure tuning のゼロ値 field は既定値（defaultCallGap / defaultRetryBackoffBase / defaultRetryBackoffMax）へフォールバックする。
// @ensure Tuning{} を渡した場合の挙動は NewSpeechSynthesizer と同一。
func NewSpeechSynthesizerWithTuning(httpClient *http.Client, apiKey string, tuning Tuning) *SpeechSynthesizer {
	s := newSpeechSynthesizerForTest(httpClient, apiKey, time.Sleep)
	s.callGap = firstNonZeroDuration(tuning.CallGap, defaultCallGap)
	s.retryBackoffBase = firstNonZeroDuration(tuning.RetryBackoffBase, defaultRetryBackoffBase)
	s.retryBackoffMax = firstNonZeroDuration(tuning.RetryBackoffMax, defaultRetryBackoffMax)
	return s
}

func firstNonZeroDuration(v, fallback time.Duration) time.Duration {
	if v == 0 {
		return fallback
	}
	return v
}

func newSpeechSynthesizerForTest(httpClient *http.Client, apiKey string, backoffSleepFn func(time.Duration)) *SpeechSynthesizer {
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	return &SpeechSynthesizer{
		client:           httpClient,
		apiKey:           apiKey,
		backoffSleepFn:   backoffSleepFn,
		nowFn:            time.Now,
		callGap:          defaultCallGap,
		retryBackoffBase: defaultRetryBackoffBase,
		retryBackoffMax:  defaultRetryBackoffMax,
	}
}
