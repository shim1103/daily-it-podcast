package constants_test

import (
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// TestTextWriterBriefPrompt_hasAllNumericPlaceholders は、brief template が
// build.embedManuscriptDraftLimits が埋める数値 placeholder 一式を漏れなく含むことを固定する。
// prompt variant（testdata/brief_prompt_variant_*.txt）も同じ placeholder 集合を保つ前提。
func TestTextWriterBriefPrompt_hasAllNumericPlaceholders(t *testing.T) {
	t.Parallel()

	required := []string{
		"{{TITLE_MIN}}", "{{TITLE_MAX}}", "{{TITLE_TARGET}}",
		"{{INTRO_MIN}}", "{{INTRO_MAX}}", "{{INTRO_TARGET}}",
		"{{CLOSING_MIN}}", "{{CLOSING_MAX}}", "{{CLOSING_TARGET}}",
		"{{TOPIC_TITLE_MIN}}", "{{TOPIC_TITLE_MAX}}", "{{TOPIC_TITLE_TARGET}}",
		"{{PREFACE_MIN}}", "{{PREFACE_MAX}}", "{{PREFACE_TARGET}}",
		"{{DETAIL_MIN}}", "{{DETAIL_MAX}}", "{{DETAIL_TARGET}}",
		"{{TOPIC_COUNT_MIN}}", "{{TOPIC_COUNT_MAX}}", "{{TOPIC_COUNT_TARGET}}",
		"{{TOTAL_MIN}}", "{{TOTAL_MAX}}", "{{TOTAL_TARGET}}",
		"{{TOTAL_MINUTES_MIN}}", "{{TOTAL_MINUTES_MAX}}",
	}

	for _, ph := range required {
		if !strings.Contains(constants.TextWriterBriefPrompt, ph) {
			t.Errorf("TextWriterBriefPrompt に数値 placeholder %s が無い", ph)
		}
	}
}

// TestTextWriterBriefPrompt_hasDynamicPlaceholders は {{SOURCES}} / {{JSON_EXAMPLE}} の存在を固定する。
func TestTextWriterBriefPrompt_hasDynamicPlaceholders(t *testing.T) {
	t.Parallel()

	for _, ph := range []string{"{{SOURCES}}", "{{JSON_EXAMPLE}}"} {
		if !strings.Contains(constants.TextWriterBriefPrompt, ph) {
			t.Errorf("TextWriterBriefPrompt に動的 placeholder %s が無い", ph)
		}
	}
}

// TestTextWriterBriefPrompt_capsShortFieldsHard は、上限を超えやすい短い field
// （title / intro / closingSummary）に対し「上限を絶対に超えない」旨と、収まる文の本数の
// 目安を与える指導文があることを固定する（Gemini が intro / closingSummary を上限超過、
// title を下限割れさせた実測 run 34202239672 への対応。validation はしない）。
func TestTextWriterBriefPrompt_capsShortFieldsHard(t *testing.T) {
	t.Parallel()

	p := constants.TextWriterBriefPrompt
	required := []string{
		"# Short-field length（最優先で厳守）",
		"上限文字数を 1 文字でも超えたら不合格",
		"intro と closingSummary は 3 文以内・各文 45 文字以内",
		"title は 1 行の見出しで、下限文字数を必ず満たす",
		// 実測で preface が 66 文字（下限 70 割れ）に落ちたため、下限を明示する。
		"各 topic.preface は {{PREFACE_MIN}} 文字未満なら不合格。2〜3 文で書き、短ければ背景を 1 文足す",
		"書き終えたら文字数を数え、超過していれば文を削って上限内へ必ず収める",
		// 実測で title が 26〜29 文字に張り付いたため、具体的な語数アンカーを置く。
		"{{TITLE_MIN}} 文字は日本語で 20 字前後の見出しに説明句を 1 つ足した長さ",
	}
	for _, s := range required {
		if !strings.Contains(p, s) {
			t.Errorf("TextWriterBriefPrompt に %q が無い", s)
		}
	}
}

// TestTextWriterBriefPrompt_hasMandatoryFixStep は、提出前に各 field の文字数を数えて
// range 外なら機械的に足す／削る「必須修正手順」を持つことを固定する（Gemini が range 際で
// 張り付く実測 run 34203032590 への対応）。
func TestTextWriterBriefPrompt_hasMandatoryFixStep(t *testing.T) {
	t.Parallel()

	p := constants.TextWriterBriefPrompt
	required := []string{
		"# 提出前の必須修正手順",
		"下限より少なければ語句を足し、上限より多ければ語句を削る",
		"この修正を全 field が range 内に収まるまで繰り返してから出力する",
	}
	for _, s := range required {
		if !strings.Contains(p, s) {
			t.Errorf("TextWriterBriefPrompt に %q が無い", s)
		}
	}
}

// TestTextWriterBriefPrompt_enforcesTotalMinimumWithRecipe は、全体合計の下限割れを防ぐため
// 具体的な文字数配分の目安（topic 5 件・各 detail の下限、合計が下限を割ったら detail を伸ばす）
// を Length strategy に持つことを固定する（Gemini が total 下限割れした実測 run 34202649569 への対応）。
func TestTextWriterBriefPrompt_enforcesTotalMinimumWithRecipe(t *testing.T) {
	t.Parallel()

	p := constants.TextWriterBriefPrompt
	required := []string{
		"合計が {{TOTAL_MIN}} 文字を下回ったら不合格",
		"各 detail は {{DETAIL_TARGET}} 文字を目標に書く（{{DETAIL_MIN}} は最低ライン、通常はそれより長く書く）",
		"合計を数えて {{TOTAL_MIN}} に満たなければ、各 detail に説明を足して {{TOTAL_MIN}} 文字以上へ必ず伸ばす",
	}
	for _, s := range required {
		if !strings.Contains(p, s) {
			t.Errorf("TextWriterBriefPrompt に %q が無い", s)
		}
	}
}

// TestTextWriterBriefPrompt_guidesDetailNewlines_whenParagraphBreakNeeded は
// 改行を入れてよいのは topic.detail の中だけ・detail の段落は1個だけ・他 field は改行禁止の
// 指導文があることを固定する（validation はしない）。
func TestTextWriterBriefPrompt_guidesDetailNewlines_whenParagraphBreakNeeded(t *testing.T) {
	t.Parallel()

	p := constants.TextWriterBriefPrompt
	required := []string{
		"# topic detail",
		"改行を入れてよいのは topic.detail の中だけ",
		"段落は 1 個だけ",
		"無意味な空白",
		"title / intro / topic.title / topic.preface / closingSummary には改行を一切入れない",
	}
	for _, s := range required {
		if !strings.Contains(p, s) {
			t.Errorf("TextWriterBriefPrompt に %q が無い", s)
		}
	}
}
