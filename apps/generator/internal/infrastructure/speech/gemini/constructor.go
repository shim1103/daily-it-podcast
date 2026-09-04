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

// withCallTimeout は httpClient の shallow copy に Gemini TTS 1 呼び出しぶんの
// 全体 timeout（httpCallTimeout）を付けて返す。
// why: Composition が渡す *http.Client は全体 timeout を持たない。1 呼び出しの上限は
//
//	vendor 固有制約なので Adapter が付け直す。呼び出し元の Client は変更しない。
//
// @ensure httpClient == nil のときは nil を返す（nil client の防御は SynthesizeAll が持つ）。
// @ensure 非 nil のときは httpCallTimeout を持つ別 *http.Client。引数の Client は変更しない。
func withCallTimeout(httpClient *http.Client) *http.Client {
	if httpClient == nil {
		return nil
	}
	c := *httpClient
	c.Timeout = httpCallTimeout
	return &c
}

func newSpeechSynthesizerForTest(httpClient *http.Client, apiKey string, backoffSleepFn func(time.Duration)) *SpeechSynthesizer {
	if backoffSleepFn == nil {
		backoffSleepFn = time.Sleep
	}
	return &SpeechSynthesizer{
		client:           withCallTimeout(httpClient),
		apiKey:           apiKey,
		backoffSleepFn:   backoffSleepFn,
		nowFn:            time.Now,
		callGap:          defaultCallGap,
		retryBackoffBase: defaultRetryBackoffBase,
		retryBackoffMax:  defaultRetryBackoffMax,
	}
}
