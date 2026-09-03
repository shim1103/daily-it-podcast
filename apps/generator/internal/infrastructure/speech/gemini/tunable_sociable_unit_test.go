package gemini

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestSpeechSynthesizer_usesInjectedCallGap_whenTuningProvided は
// NewSpeechSynthesizerWithTuning で短い CallGap を注入すると、連続 Synthesize の
// 待機が既定 20s ではなく注入値どおりに縮むことを検証する。
func TestSpeechSynthesizer_usesInjectedCallGap_whenTuningProvided(t *testing.T) {
	// Given: 常に成功応答を返す fake client と、待機呼び出しを記録する sleep spy
	rt := &fakeRoundTripper{responses: []fakeClientResponse{
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
	}}
	var sleeps []time.Duration
	const injectedGap = 5 * time.Millisecond
	synth := NewSpeechSynthesizerWithTuning(&http.Client{Transport: rt}, "gemini-fake-key", Tuning{CallGap: injectedGap})
	synth.backoffSleepFn = func(d time.Duration) { sleeps = append(sleeps, d) }
	// nowFn は常に同じ時刻。elapsed=0 なので毎回 callGap 全量の待機が入るはず。
	synth.nowFn = func() time.Time { return time.Unix(0, 0) }

	// When: 2 回連続で Synthesize する
	for i := 0; i < 2; i++ {
		if _, err := synth.Synthesize(context.Background(), "注入テスト"); err != nil {
			t.Fatalf("Synthesize(%d): %v", i+1, err)
		}
	}

	// Then: 待機は 1 度だけ（2 回目の呼び出し前）で、注入値どおり。既定 20s ではない。
	if len(sleeps) != 1 {
		t.Fatalf("sleep 呼び出し回数 = %d (%v), want 1", len(sleeps), sleeps)
	}
	if sleeps[0] != injectedGap {
		t.Fatalf("sleep = %v, want %v（注入 CallGap）", sleeps[0], injectedGap)
	}
}

// TestSpeechSynthesizer_skipsCallGapWait_whenElapsedExceedsInjectedGap は
// 前回呼び出しからの経過が注入 CallGap を超えていれば待機しないことを検証する。
func TestSpeechSynthesizer_skipsCallGapWait_whenElapsedExceedsInjectedGap(t *testing.T) {
	// Given: 成功応答 2 回分の fake client
	rt := &fakeRoundTripper{responses: []fakeClientResponse{
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
		{status: http.StatusOK, body: jsonBody(t, audioInteractionResponse(minimalPCM()))},
	}}
	var sleeps []time.Duration
	synth := NewSpeechSynthesizerWithTuning(&http.Client{Transport: rt}, "gemini-fake-key", Tuning{CallGap: time.Second})
	synth.backoffSleepFn = func(d time.Duration) { sleeps = append(sleeps, d) }
	// nowFn は呼ぶたびに 10s 進む。経過 >> CallGap なので待機は入らない。
	base := time.Unix(0, 0)
	calls := 0
	synth.nowFn = func() time.Time {
		calls++
		return base.Add(time.Duration(calls) * 10 * time.Second)
	}

	// When: 2 回連続で Synthesize する
	for i := 0; i < 2; i++ {
		if _, err := synth.Synthesize(context.Background(), "経過超過テスト"); err != nil {
			t.Fatalf("Synthesize(%d): %v", i+1, err)
		}
	}

	// Then: callGap 待機は発生しない
	if len(sleeps) != 0 {
		t.Fatalf("sleep 呼び出し = %v, want なし（経過が CallGap を超過）", sleeps)
	}
}

// TestSpeechSynthesizer_fallsBackToDefaultTuning_whenZeroValueFields は
// Tuning のゼロ値 field が既定値へフォールバックすることを検証する。
func TestSpeechSynthesizer_fallsBackToDefaultTuning_whenZeroValueFields(t *testing.T) {
	// Given: 空 Tuning を注入した Synthesizer
	synth := NewSpeechSynthesizerWithTuning(&http.Client{}, "gemini-fake-key", Tuning{})

	// Then: 各 tuning field が既定値
	if synth.callGap != defaultCallGap {
		t.Fatalf("callGap = %v, want %v", synth.callGap, defaultCallGap)
	}
	if synth.retryBackoffBase != defaultRetryBackoffBase {
		t.Fatalf("retryBackoffBase = %v, want %v", synth.retryBackoffBase, defaultRetryBackoffBase)
	}
	if synth.retryBackoffMax != defaultRetryBackoffMax {
		t.Fatalf("retryBackoffMax = %v, want %v", synth.retryBackoffMax, defaultRetryBackoffMax)
	}
}

// TestNewSpeechSynthesizer_usesDefaultTuning_whenConstructedPlainly は
// 既定 constructor が既定 tuning を使う（挙動不変）ことを検証する。
func TestNewSpeechSynthesizer_usesDefaultTuning_whenConstructedPlainly(t *testing.T) {
	// Given / When: 既定 constructor
	synth := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key")

	// Then: tuning field はすべて既定値
	if synth.callGap != defaultCallGap || synth.retryBackoffBase != defaultRetryBackoffBase || synth.retryBackoffMax != defaultRetryBackoffMax {
		t.Fatalf("tuning = {%v, %v, %v}, want defaults {%v, %v, %v}",
			synth.callGap, synth.retryBackoffBase, synth.retryBackoffMax,
			defaultCallGap, defaultRetryBackoffBase, defaultRetryBackoffMax)
	}
}
