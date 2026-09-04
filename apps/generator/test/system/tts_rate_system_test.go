//go:build system && ratemeasure

// Scope: System（Gemini TTS の PASS 率・所要計測。dispatch 専用）
// 実物: gemini.SpeechSynthesizer が実 GEMINI_API_KEY で実 Gemini TTS を叩く。
// Double: なし。Cursor / GetX / Drive は呼ばない。
// 目的: build.SpeechTexts が返す本番 topic 束の Max 相当（preface+detail を Domain 上限いっぱい）を
//
//	runs 回直列 SynthesizeAll し、実 Gemini 応答が err == nil で返る率が閾値以上かを計測する
//	（Decision 2026-09-03T14-46-00）。
//	短文（greeting+intro 束）では長尺 TTS 特有の失敗（MAX_TOKENS truncation / audio 途切れ /
//	長 base64 decode）が観測できないため、本番の最長経路（DraftTopicDetailMaxSec 相当）を叩く。
//	Adapter が非空・最小尺（minPCMBytes）の WAV を contract として保証するので、
//	PASS 判定は err == nil のみ（非空 WAV / 尺 > 0 の再確認は Adapter の責務なのでしない）。
//	callGap / backoff は env から注入して差し替える。
//
// @require process env に TEST_GEMINI_API_KEY がある（無ければ Skip）。本番 env 名（config.GeminiAPIKeyEnv）は読まない。
// @ensure 計測対象の束は本番 topic 束 Max の 8 割以上かつ Max（+連結改行）以下。外れれば t.Fatalf。
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

// jpFiller は maxRunes を超えない範囲で文を最大限繰り返し、本番相当の朗読 text を組む。
// 末尾は必ず句点で終える（Domain の文末規則に合わせる）。maxRunes 未満で 1 文も入らない
// ことはない前提（呼び出し側が Domain の *MaxLen を渡す）。
func jpFiller(maxRunes int, sentence string) string {
	var b strings.Builder
	for utf8.RuneCountInString(b.String()+sentence) <= maxRunes {
		b.WriteString(sentence)
	}
	s := b.String()
	if s == "" { // maxRunes が 1 文より短い異常時の保険。
		s = sentence
	}
	if !strings.HasSuffix(s, "。") {
		s += "。"
	}
	return s
}

// rateMeasureDraft は本番の Domain 上限いっぱいの長尺の疑似原稿を返す。
// preface / detail は DraftTopicPrefaceMaxLen / DraftTopicDetailMaxLen を超えない最大長にして、
// SpeechTexts が返す topic 束が本番 topic 束の Max 相当（最長経路）の再生尺を持つようにする。
func rateMeasureDraft() models.ManuscriptDraft {
	const prefaceSentence = "この話題では、発表の背景と要点をかいつまんで先にお伝えします。"
	const detailSentence = "詳細としては、公表された内容とその影響範囲、想定される利用場面、既存の仕組みとの違い、そして今後の見通しについて順に説明していきます。"
	preface := jpFiller(constants.DraftTopicPrefaceMaxLen, prefaceSentence)
	detail := jpFiller(constants.DraftTopicDetailMaxLen, detailSentence)
	return models.ManuscriptDraft{
		Title:          "きょうの IT ニュースまとめ回",
		Intro:          jpFiller(constants.DraftIntroMaxLen, "本日は注目の発表をまとめてお届けします。"),
		ClosingSummary: jpFiller(constants.DraftClosingMaxLen, "本日取り上げた話題を振り返ります。"),
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

	// Given: 本番相当の長尺原稿から SpeechTexts が返す topic 束（texts[1] = preface+detail の最長経路）
	texts := build.SpeechTexts("こんにちは。", "さようなら。", rateMeasureDraft())
	if len(texts) < 3 {
		t.Fatalf("SpeechTexts 本数 = %d, want >= 3（greeting+intro / topic / closing+farewell）", len(texts))
	}
	head := texts[1]
	headRunes := utf8.RuneCountInString(head)

	// Then（precondition）: 計測対象は本番 topic 束の Max 相当（最長経路）。
	// 長尺 TTS 特有の失敗を狙うため、Max の 8 割以上かつ Max（+ 連結改行）を超えない範囲に収める。
	maxRunes := constants.DraftTopicPrefaceMaxLen + constants.DraftTopicDetailMaxLen + 1 // +1 は preface と detail の連結改行
	floorRunes := maxRunes * 8 / 10
	if headRunes < floorRunes {
		t.Fatalf("計測対象の束 = %d rune, want >= %d（本番 topic 束 Max %d の 8 割）", headRunes, floorRunes, maxRunes)
	}
	if headRunes > maxRunes {
		t.Fatalf("計測対象の束 = %d rune, want <= %d（本番 topic 束 Max を超えない）", headRunes, maxRunes)
	}
	t.Logf("計測対象の束: %d rune（本番 topic 束 Max %d 相当。目安 %.0fs）", headRunes, maxRunes,
		float64(headRunes)/float64(constants.CharsPerSecond))

	// When / Then: 本番相当の topic 束を runs 回直列 SynthesizeAll（単一 text）。err == nil で PASS。
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
