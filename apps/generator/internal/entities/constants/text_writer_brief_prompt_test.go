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
		"intro と closingSummary は 3〜5 文",
		"title は 1 行の見出しで、下限文字数を必ず満たす",
		"書き終えたら文字数を数え、超過していれば文を削って上限内へ必ず収める",
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
