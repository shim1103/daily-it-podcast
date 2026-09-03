//go:build system

// Scope: System（TTS 単体到達）
// 実物: gemini.SpeechSynthesizer が実 GEMINI_API_KEY で実 Gemini TTS を叩く。
// Double: なし。Cursor / GetX / Drive は呼ばない。
// 目的: build.SpeechTexts が返す topic+2 束の朗読 text を順に Synthesize し、
//
//	TEST key の rate 制約（RPD / RPM）内で完走できることを ensure する（Decision 2026-09-02T13-57-00 / 2026-09-03T14-45-00）。
//	429 到達は「TEST key の rate 制約に収まらなかった」失敗として扱う。
//
// @require process env に GEMINI_API_KEY（TEST key）がある（無ければ Skip）。full env は要らない。
// @ensure topic+2 回の Synthesize がすべて非空 WAV（RIFF/WAVE）を返し、build.WavDurationSec が正の秒数を返す。
// @invariant local に secret を置かない。Drive へ書かない。cleanup 不要。
package system

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
)

// pseudoDraft は TTS 到達 test 用の疑似原稿（topic 3）。
// 各 field は短い固定日本語文。Domain 定数の下限は満たさなくてよい（draft 検証は通さない）。
func pseudoDraft() models.ManuscriptDraft {
	return models.ManuscriptDraft{
		Title:          "きょうの IT ニュース",
		Intro:          "本日の導入です。",
		ClosingSummary: "本日のまとめです。",
		Topics: []models.ManuscriptDraftTopic{
			{Title: "話題一", Preface: "一つ目の前置きです。", Detail: "一つ目の詳細です。"},
			{Title: "話題二", Preface: "二つ目の前置きです。", Detail: "二つ目の詳細です。"},
			{Title: "話題三", Preface: "三つ目の前置きです。", Detail: "三つ目の詳細です。"},
		},
	}
}

func TestSpeechSynthesisSystem_synthesizesTopicPlusTwoBundlesWithinFreeQuota_whenRealAPIKeyPresent(t *testing.T) {
	// Given: 実 GEMINI_API_KEY（無ければ Skip）で組んだ SpeechSynthesizer
	apiKey := strings.TrimSpace(os.Getenv(config.GeminiAPIKeyEnv))
	if apiKey == "" {
		t.Skipf("System precondition: %s が無い（TTS 到達 test を skip）", config.GeminiAPIKeyEnv)
	}
	synth := gemini.NewSpeechSynthesizer(&http.Client{Timeout: 5 * time.Minute}, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Given: 疑似原稿（topic 3）から topic+2 束の朗読 text
	d := pseudoDraft()
	texts := build.SpeechTexts("こんにちは。", "さようなら。", d)
	wantCount := 1 + len(d.Topics) + 1
	if len(texts) != wantCount {
		t.Fatalf("SpeechTexts 本数 = %d, want %d", len(texts), wantCount)
	}

	// When / Then: 各束を順に Synthesize。すべて非空 WAV かつ尺 > 0
	for i, text := range texts {
		audio, err := synth.Synthesize(ctx, text)
		if err != nil {
			if strings.Contains(err.Error(), "status 429") {
				t.Fatalf("無料枠 quota を超えた（%d/%d 本目で 429）: %v", i+1, len(texts), err)
			}
			t.Fatalf("Synthesize(%d/%d): %v", i+1, len(texts), err)
		}
		if !isWAV(audio.Content) {
			head := audio.Content
			if len(head) > 12 {
				head = head[:12]
			}
			t.Fatalf("Synthesize(%d/%d): 非 WAV 応答 head=% x", i+1, len(texts), head)
		}
		dur, err := build.WavDurationSec(audio.Content)
		if err != nil {
			t.Fatalf("WavDurationSec(%d/%d): %v", i+1, len(texts), err)
		}
		if dur <= 0 {
			t.Fatalf("WavDurationSec(%d/%d) = %v, want > 0", i+1, len(texts), dur)
		}
	}
}

// isWAV は bytes が RIFF/WAVE ヘッダを持つかを返す。
func isWAV(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'A' && data[10] == 'V' && data[11] == 'E'
}
