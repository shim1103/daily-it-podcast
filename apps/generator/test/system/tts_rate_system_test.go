//go:build system && ratemeasure

// Scope: System（Gemini TTS の PASS 率・所要計測。dispatch 専用）
// 実物: gemini.SpeechSynthesizer が実 GEMINI_API_KEY で実 Gemini TTS を叩く。
// Double: なし。Cursor / GetX / Drive は呼ばない。
// 目的: build.SpeechTexts の先頭 1 束だけを runs 回直列 Synthesize し、
//
//	実 Gemini 応答が err == nil で返る率が閾値以上かを計測する（Decision 2026-09-03T14-46-00）。
//	Adapter が非空・最小尺（minPCMBytes）の WAV を contract として保証するので、
//	PASS 判定は err == nil のみ（非空 WAV / 尺 > 0 の再確認は Adapter の責務なのでしない）。
//	callGap / backoff は env から注入して差し替える。
//
// @require process env に TEST_GEMINI_API_KEY がある（無ければ Skip）。本番 env 名（config.GeminiAPIKeyEnv）は読まない。
// @ensure pass/runs >= pass_threshold で緑、下回れば t.Fatalf。各回の所要秒と PASS 率・tuning 値を Logf。
// @invariant 既定 -tags=system では compile されない（ratemeasure tag）。local に secret を置かない。本番 key が計測へ流れない（TEST_ 直読み）。
package system

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

func ttsEnvInt(t *testing.T, key string, def int) int {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		t.Fatalf("%s = %q, want 正の整数", key, raw)
	}
	return v
}

func ttsEnvDuration(t *testing.T, key string, def time.Duration) time.Duration {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		t.Fatalf("%s = %q, want time.ParseDuration 可能な正の値", key, raw)
	}
	return d
}

func ttsEnvFloat(t *testing.T, key string, def float64) float64 {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 || v > 1 {
		t.Fatalf("%s = %q, want (0,1] の小数", key, raw)
	}
	return v
}

func TestGeminiTTSRate_measuresPassRate_overNRuns(t *testing.T) {
	// Given: 実 TEST_GEMINI_API_KEY（無ければ Skip）。本番 env 名は読まない（Decision 2026-09-03T16-30-00）。
	const geminiAPIKeyEnv = "TEST_GEMINI_API_KEY"
	apiKey := strings.TrimSpace(os.Getenv(geminiAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("計測 precondition: %s が無い（TTS rate 計測を skip）", geminiAPIKeyEnv)
	}

	runs := ttsEnvInt(t, "TTS_RATE_RUNS", 10)
	callGap := ttsEnvDuration(t, "TTS_CALL_GAP", 20*time.Second)
	backoffBase := ttsEnvDuration(t, "TTS_RETRY_BACKOFF_BASE", 60*time.Second)
	backoffMax := ttsEnvDuration(t, "TTS_RETRY_BACKOFF_MAX", 3*time.Minute)
	passThreshold := ttsEnvFloat(t, "TTS_PASS_THRESHOLD", 0.8)

	synth := gemini.NewSpeechSynthesizerWithTuning(
		&http.Client{Timeout: 5 * time.Minute},
		apiKey,
		gemini.Tuning{CallGap: callGap, RetryBackoffBase: backoffBase, RetryBackoffMax: backoffMax},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	// Given: 疑似原稿の先頭 1 束の朗読 text（speech_synthesis_system_test.go の pseudoDraft を流用）
	texts := build.SpeechTexts("こんにちは。", "さようなら。", pseudoDraft())
	if len(texts) == 0 {
		t.Fatal("SpeechTexts が空")
	}
	head := texts[0]

	// When / Then: 先頭 1 束を runs 回直列 SynthesizeAll（単一 text）。err == nil で PASS。
	// Adapter が非空・最小尺の WAV を contract として保証するので、ここでは再確認しない。
	pass := 0
	var durations []float64
	for i := 1; i <= runs; i++ {
		start := time.Now()
		_, err := synth.SynthesizeAll(ctx, []string{head})
		elapsed := time.Since(start).Seconds()
		durations = append(durations, elapsed)
		if err != nil {
			t.Logf("run %d/%d: FAIL（%v）所要 %.1fs", i, runs, err, elapsed)
			continue
		}
		pass++
		t.Logf("run %d/%d: PASS 所要 %.1fs", i, runs, elapsed)
	}

	var avg float64
	for _, d := range durations {
		avg += d
	}
	if len(durations) > 0 {
		avg /= float64(len(durations))
	}
	rate := float64(pass) / float64(runs)
	t.Logf("PASS率 %d/%d 平均 %.1fs callGap=%s backoffBase=%s backoffMax=%s", pass, runs, avg, callGap, backoffBase, backoffMax)

	if rate < passThreshold {
		t.Fatalf("PASS率 %.2f < 閾値 %.2f（%d/%d）", rate, passThreshold, pass, runs)
	}
}
