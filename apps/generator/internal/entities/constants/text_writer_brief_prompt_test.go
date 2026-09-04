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
