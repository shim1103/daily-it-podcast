package build

import (
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

func TestEmbedManuscriptDraftLimits_全placeholderを定数で置換する(t *testing.T) {
	t.Parallel()

	// Given: manuscript_draft_limits の全 placeholder が含まれる template
	template := strings.Join([]string{
		"{{TITLE_MIN}} {{TITLE_MAX}} {{TITLE_TARGET}}",
		"{{INTRO_MIN}} {{INTRO_MAX}} {{INTRO_TARGET}}",
		"{{TOPIC_TITLE_MIN}} {{TOPIC_TITLE_MAX}} {{TOPIC_TITLE_TARGET}}",
		"{{PREFACE_MIN}} {{PREFACE_MAX}} {{PREFACE_TARGET}}",
		"{{DETAIL_MIN}} {{DETAIL_MAX}} {{DETAIL_TARGET}}",
		"{{TOPIC_COUNT_MIN}} {{TOPIC_COUNT_MAX}} {{TOPIC_COUNT_TARGET}}",
		"{{TOTAL_MIN}} {{TOTAL_MAX}} {{TOTAL_TARGET}}",
	}, " ")

	// When: 数値 placeholder を埋める
	got := embedManuscriptDraftLimits(template)

	// Then: placeholder が残らない
	if strings.Contains(got, "{{") || strings.Contains(got, "}}") {
		t.Fatalf("embedManuscriptDraftLimits: placeholder が残っている: %q", got)
	}

	// Then: TextWriterBriefPrompt の数値 placeholder はすべて埋まる
	prompt := embedManuscriptDraftLimits(constants.TextWriterBriefPrompt)
	for _, ph := range []string{
		"{{TITLE_MIN}}", "{{TOTAL_TARGET}}", "{{PREFACE_MAX}}",
	} {
		if strings.Contains(prompt, ph) {
			t.Fatalf("TextWriterBriefPrompt: %s が残っている", ph)
		}
	}

	// Then: ComposeBrief 担当の動的 placeholder は残る
	for _, ph := range []string{"{{SOURCES}}", "{{JSON_EXAMPLE}}"} {
		if !strings.Contains(prompt, ph) {
			t.Fatalf("TextWriterBriefPrompt: %s が消えている", ph)
		}
	}
}
