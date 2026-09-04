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
// @ensure 計測対象の束は TTS_DOUBLE の尺帯（profile 目標長の 7 割以上かつ 目標長+連結改行 以下）。外れれば t.Fatalf。
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
// 末尾は必ず句点で終える（Domain の文末規則に合わせる）。targetRunes 未満で 1 文も入らない
// ことはない前提（呼び出し側が Domain の *Len を渡す）。
func jpFiller(targetRunes int, sentence string) string {
	var b strings.Builder
	for utf8.RuneCountInString(b.String()+sentence) <= targetRunes {
		b.WriteString(sentence)
	}
	s := b.String()
	if s == "" { // targetRunes が 1 文より短い異常時の保険。
		s = sentence
	}
	if !strings.HasSuffix(s, "。") {
		s += "。"
	}
	return s
}

// ttsDoubleTarget は計測に使う朗読尺プロファイル。TTS_DOUBLE（min / tgt / max）で切り替える。
// 本番の topic 束の再生尺帯（Domain の *MinLen / *Target / *MaxLen）に対応する。
type ttsDoubleTarget struct {
	name       string
	prefaceLen int
	detailLen  int
	introLen   int
	closingLen int
}

// ttsDoubleProfile は TTS_DOUBLE の env 値から尺プロファイルを引く。
// 未知値・空値は max（最長経路）へフォールバックする（誤字で測り漏れる側に倒さない）。
func ttsDoubleProfile(raw string) ttsDoubleTarget {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "min":
		return ttsDoubleTarget{
			name:       "min",
			prefaceLen: constants.DraftTopicPrefaceMinLen,
			detailLen:  constants.DraftTopicDetailMinLen,
			introLen:   constants.DraftIntroMinLen,
			closingLen: constants.DraftClosingMinLen,
		}
	case "tgt", "target":
		return ttsDoubleTarget{
			name:       "tgt",
			prefaceLen: constants.DraftTopicPrefaceTarget,
			detailLen:  constants.DraftTopicDetailTarget,
			introLen:   constants.DraftIntroTarget,
			closingLen: constants.DraftClosingTarget,
		}
	default: // "" / "max" / 未知値
		return ttsDoubleTarget{
			name:       "max",
			prefaceLen: constants.DraftTopicPrefaceMaxLen,
			detailLen:  constants.DraftTopicDetailMaxLen,
			introLen:   constants.DraftIntroMaxLen,
			closingLen: constants.DraftClosingMaxLen,
		}
	}
}

// rateMeasureDraft は指定プロファイルの尺を狙った長尺の疑似原稿を返す。
// preface / detail は profile の目標長を超えない最大長にして、SpeechTexts が返す topic 束が
// その帯（min / tgt / max）の再生尺を持つようにする。
func rateMeasureDraft(p ttsDoubleTarget) models.ManuscriptDraft {
	const prefaceSentence = "この話題では、発表の背景と要点をかいつまんで先にお伝えします。"
	const detailSentence = "詳細としては、公表された内容とその影響範囲、想定される利用場面、既存の仕組みとの違い、そして今後の見通しについて順に説明していきます。"
	preface := jpFiller(p.prefaceLen, prefaceSentence)
	detail := jpFiller(p.detailLen, detailSentence)
	return models.ManuscriptDraft{
		Title:          "きょうの IT ニュースまとめ回",
		Intro:          jpFiller(p.introLen, "本日は注目の発表をまとめてお届けします。"),
		ClosingSummary: jpFiller(p.closingLen, "本日取り上げた話題を振り返ります。"),
		Topics: []models.ManuscriptDraftTopic{
			{Title: "話題一の見出しです", Preface: preface, Detail: detail},
			{Title: "話題二の見出しです", Preface: preface, Detail: detail},
			{Title: "話題三の見出しです", Preface: preface, Detail: detail},
		},
	}
}

func TestTTSDoubleProfile_selectsBundleTargetLens_fromEnvValue(t *testing.T) {
	cases := []struct {
		env         string
		wantPreface int
		wantDetail  int
	}{
		{"", constants.DraftTopicPrefaceMaxLen, constants.DraftTopicDetailMaxLen},      // 既定は max
		{"max", constants.DraftTopicPrefaceMaxLen, constants.DraftTopicDetailMaxLen},   // 明示 max
		{"MAX", constants.DraftTopicPrefaceMaxLen, constants.DraftTopicDetailMaxLen},   // 大文字ゆらぎ
		{"tgt", constants.DraftTopicPrefaceTarget, constants.DraftTopicDetailTarget},   // 目標
		{" tgt ", constants.DraftTopicPrefaceTarget, constants.DraftTopicDetailTarget}, // 前後空白
		{"min", constants.DraftTopicPrefaceMinLen, constants.DraftTopicDetailMinLen},   // 下限
	}
	for _, c := range cases {
		p := ttsDoubleProfile(c.env)
		if p.prefaceLen != c.wantPreface || p.detailLen != c.wantDetail {
			t.Fatalf("ttsDoubleProfile(%q) = {preface %d, detail %d}, want {%d, %d}",
				c.env, p.prefaceLen, p.detailLen, c.wantPreface, c.wantDetail)
		}
	}
}

func TestTTSDoubleProfile_fallsBackToMax_whenUnknownValue(t *testing.T) {
	// 未知値は既定（max）へフォールバックする。誤字で黙って別尺にならないよう、
	// フォールバックしても既定と同じ最長経路にはなる（測り漏れを防ぐ側に倒す）。
	p := ttsDoubleProfile("huge")
	if p.prefaceLen != constants.DraftTopicPrefaceMaxLen || p.detailLen != constants.DraftTopicDetailMaxLen {
		t.Fatalf("unknown value fallback = {preface %d, detail %d}, want max {%d, %d}",
			p.prefaceLen, p.detailLen, constants.DraftTopicPrefaceMaxLen, constants.DraftTopicDetailMaxLen)
	}
}

func TestRateMeasureDraft_topicBundleStaysWithinProfileBand_forEachDouble(t *testing.T) {
	for _, env := range []string{"min", "tgt", "max"} {
		p := ttsDoubleProfile(env)
		texts := build.SpeechTexts("こんにちは。", "さようなら。", rateMeasureDraft(p))
		if len(texts) < 3 {
			t.Fatalf("%s: SpeechTexts 本数 = %d, want >= 3", env, len(texts))
		}
		got := utf8.RuneCountInString(texts[1])
		// 束の上限は preface+detail の目標長 + 連結改行 + 1 文ぶんの繰り返し許容。
		ceil := p.prefaceLen + p.detailLen + 1
		floor := ceil * 7 / 10
		if got < floor || got > ceil {
			t.Fatalf("%s: topic 束 = %d rune, want %d..%d", env, got, floor, ceil)
		}
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
	profile := ttsDoubleProfile(os.Getenv("TTS_DOUBLE"))

	synth := gemini.NewSpeechSynthesizerWithTuning(
		&http.Client{Timeout: 5 * time.Minute},
		apiKey,
		gemini.Tuning{CallGap: callGap, RetryBackoffBase: backoffBase, RetryBackoffMax: backoffMax},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	// Given: profile（min / tgt / max）の尺を狙った原稿から SpeechTexts が返す topic 束（texts[1] = preface+detail）
	texts := build.SpeechTexts("こんにちは。", "さようなら。", rateMeasureDraft(profile))
	if len(texts) < 3 {
		t.Fatalf("SpeechTexts 本数 = %d, want >= 3（greeting+intro / topic / closing+farewell）", len(texts))
	}
	head := texts[1]
	headRunes := utf8.RuneCountInString(head)

	// Then（precondition）: 計測対象は profile の帯に収まる。誤って短文・過長へ退行したら落とす。
	ceilRunes := profile.prefaceLen + profile.detailLen + 1 // +1 は preface と detail の連結改行
	floorRunes := ceilRunes * 7 / 10
	if headRunes < floorRunes || headRunes > ceilRunes {
		t.Fatalf("計測対象の束 = %d rune, want %d..%d（TTS_DOUBLE=%s の帯）", headRunes, floorRunes, ceilRunes, profile.name)
	}
	t.Logf("計測対象の束: %d rune（TTS_DOUBLE=%s、帯 %d..%d。目安 %.0fs）", headRunes, profile.name, floorRunes, ceilRunes,
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
	t.Logf("PASS率 %d/%d 平均 %.1fs double=%s callGap=%s backoffBase=%s backoffMax=%s", pass, runs, avg, profile.name, callGap, backoffBase, backoffMax)

	if rate < passThreshold {
		t.Fatalf("PASS率 %.2f < 閾値 %.2f（%d/%d）", rate, passThreshold, pass, runs)
	}
}
