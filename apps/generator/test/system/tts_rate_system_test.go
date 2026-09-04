//go:build system && ratemeasure

// Scope: System（Gemini TTS の PASS 率・所要計測。dispatch 専用）
// 実物: gemini.SpeechSynthesizer が実 GEMINI_API_KEY で実 Gemini TTS を叩く。
// Double: なし。Cursor / GetX / Drive は呼ばない。
// 目的: build.SpeechTexts が返す本番 topic 束（preface+detail）を TTS_DOUBLE で選んだ尺帯
//
//	（min / tgt / max）で組み、runs 回直列 SynthesizeAll して実 Gemini 応答が err == nil で返る率が
//	閾値以上かを計測する（Decision 2026-09-03T14-46-00）。
//	短文（greeting+intro 束）では長尺 TTS 特有の失敗（MAX_TOKENS truncation / audio 途切れ /
//	長 base64 decode）が観測できないため、既定は max（最長経路 = DraftTopicDetailMaxSec 相当）。
//	Adapter が非空・最小尺（minPCMBytes）の WAV を contract として保証するので、
//	PASS 判定は err == nil のみ（非空 WAV / 尺 > 0 の再確認は Adapter の責務なのでしない）。
//	runs / callGap / backoff / pass_threshold / double の運用既定は generator-tts-rate.yml の
//	inputs.default が SSOT。test 側の第 3 引数は env 未設定時のローカル最小フォールバック。
//
// @require process env に TEST_GEMINI_API_KEY がある（無ければ Skip）。本番 env 名（config.GeminiAPIKeyEnv）は読まない。
// @require TTS_DOUBLE は空 / min / tgt / max のいずれか。未知値は t.Fatalf（誤字は即失敗させる）。
// @ensure 計測対象の束は選んだ尺帯（preface+detail 目標長の 8 割以上かつ 目標長+連結改行3 以下）。外れれば t.Fatalf。
// @ensure pass/runs >= pass_threshold で緑、下回れば t.Fatalf。各回の所要秒と PASS 率・double / tuning 値を Logf。
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
	"unicode/utf8"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
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

// jpFiller は targetRunes を超えない範囲で文を最大限繰り返し、本番相当の朗読 text を組む。
// 末尾は必ず句点で終える（Domain の文末規則に合わせる）。
func jpFiller(targetRunes int, sentence string) string {
	var b strings.Builder
	for utf8.RuneCountInString(b.String()+sentence) <= targetRunes {
		b.WriteString(sentence)
	}
	s := b.String()
	if !strings.HasSuffix(s, "。") {
		s += "。"
	}
	return s
}

// rateMeasureDraft は preface / detail / intro / closing を指定目標長で埋めた疑似原稿を返す。
// SpeechTexts が返す topic 束（preface+detail）がその尺帯の再生時間を持つようにする。
func rateMeasureDraft(prefaceLen, detailLen, introLen, closingLen int) models.ManuscriptDraft {
	const prefaceSentence = "この話題では、発表の背景と要点をかいつまんで先にお伝えします。"
	const detailSentence = "詳細としては、公表された内容とその影響範囲、想定される利用場面、既存の仕組みとの違い、そして今後の見通しについて順に説明していきます。"
	preface := jpFiller(prefaceLen, prefaceSentence)
	detail := jpFiller(detailLen, detailSentence)
	return models.ManuscriptDraft{
		Title:          "きょうの IT ニュースまとめ回",
		Intro:          jpFiller(introLen, "本日は注目の発表をまとめてお届けします。"),
		ClosingSummary: jpFiller(closingLen, "本日取り上げた話題を振り返ります。"),
		Topics: []models.ManuscriptDraftTopic{
			{Title: "話題一の見出しです", Preface: preface, Detail: detail},
			{Title: "話題二の見出しです", Preface: preface, Detail: detail},
			{Title: "話題三の見出しです", Preface: preface, Detail: detail},
		},
	}
}

func TestGeminiTTSRate_measuresPassRate_overNRuns(t *testing.T) {
	// Given: 実 TEST_GEMINI_API_KEY（無ければ Skip）。本番 env 名は読まない（Decision 2026-09-03T16-30-00）。
	const geminiAPIKeyEnv = "TEST_GEMINI_API_KEY"
	apiKey := strings.TrimSpace(os.Getenv(geminiAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("計測 precondition: %s が無い（TTS rate 計測を skip）", geminiAPIKeyEnv)
	}

	// why: 運用の既定値は generator-tts-rate.yml の inputs.default が正（SSOT）。
	//      ここの第 3 引数は env 未設定時のローカル最小フォールバックで、workflow では常に env が入る。
	runs := ttsEnvInt(t, "TTS_RATE_RUNS", 3)
	callGap := ttsEnvDuration(t, "TTS_CALL_GAP", time.Second)
	backoffBase := ttsEnvDuration(t, "TTS_RETRY_BACKOFF_BASE", time.Second)
	backoffMax := ttsEnvDuration(t, "TTS_RETRY_BACKOFF_MAX", 5*time.Second)
	passThreshold := ttsEnvFloat(t, "TTS_PASS_THRESHOLD", 0.8)

	// TTS_DOUBLE で本番 topic 束の尺帯を選ぶ。未知値は即失敗（誤字を黙って飲まない）。
	double := strings.TrimSpace(os.Getenv("TTS_DOUBLE"))
	var prefaceLen, detailLen, introLen, closingLen int
	switch double {
	case "", "max":
		double = "max"
		prefaceLen, detailLen = constants.DraftTopicPrefaceMaxLen, constants.DraftTopicDetailMaxLen
		introLen, closingLen = constants.DraftIntroMaxLen, constants.DraftClosingMaxLen
	case "tgt":
		prefaceLen, detailLen = constants.DraftTopicPrefaceTarget, constants.DraftTopicDetailTarget
		introLen, closingLen = constants.DraftIntroTarget, constants.DraftClosingTarget
	case "min":
		prefaceLen, detailLen = constants.DraftTopicPrefaceMinLen, constants.DraftTopicDetailMinLen
		introLen, closingLen = constants.DraftIntroMinLen, constants.DraftClosingMinLen
	default:
		t.Fatalf("TTS_DOUBLE = %q, want min / tgt / max のいずれか", double)
	}

	synth := gemini.NewSpeechSynthesizerWithTuning(
		&http.Client{Timeout: 5 * time.Minute},
		apiKey,
		gemini.Tuning{CallGap: callGap, RetryBackoffBase: backoffBase, RetryBackoffMax: backoffMax},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	// Given: 選んだ尺帯の原稿から SpeechTexts が返す topic 束（texts[1] = preface+detail）
	texts := build.SpeechTexts("こんにちは。", "さようなら。", rateMeasureDraft(prefaceLen, detailLen, introLen, closingLen))
	if len(texts) < 3 {
		t.Fatalf("SpeechTexts 本数 = %d, want >= 3（greeting+intro / topic / closing+farewell）", len(texts))
	}
	head := texts[1]
	headRunes := utf8.RuneCountInString(head)

	// Then（precondition）: 計測対象は選んだ尺帯に収まる。誤って短文・過長へ退行したら落とす。
	// 実 bundle は jpFiller の文単位切りで ceil の 96〜98% に入る（全帯で検算済み）。8 割を下限にする。
	ceilRunes := prefaceLen + detailLen + 3 // +3 は preface と detail の連結改行（SpeechTexts 境界）
	floorRunes := ceilRunes * 8 / 10
	if headRunes < floorRunes || headRunes > ceilRunes {
		t.Fatalf("計測対象の束 = %d rune, want %d..%d（TTS_DOUBLE=%s の帯）", headRunes, floorRunes, ceilRunes, double)
	}
	t.Logf("計測対象の束: %d rune（TTS_DOUBLE=%s、帯 %d..%d。目安 %.0fs）", headRunes, double, floorRunes, ceilRunes,
		float64(headRunes)/float64(constants.CharsPerSecond))

	// When / Then: topic 束を runs 回直列 SynthesizeAll（単一 text）。err == nil で PASS。
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
	t.Logf("PASS率 %d/%d 平均 %.1fs double=%s callGap=%s backoffBase=%s backoffMax=%s", pass, runs, avg, double, callGap, backoffBase, backoffMax)

	if rate < passThreshold {
		t.Fatalf("PASS率 %.2f < 閾値 %.2f（%d/%d）", rate, passThreshold, pass, runs)
	}
}
