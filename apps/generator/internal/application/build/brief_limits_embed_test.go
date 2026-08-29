package build

import (
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

func TestEmbedManuscriptDraftLimits_replacesNumericPlaceholdersButKeepsDynamicOnes(t *testing.T) {
	t.Parallel()

	// Given: 本番で実際に渡る TextWriterBriefPrompt（数値 placeholder と
	// 動的 placeholder {{SOURCES}} / {{JSON_EXAMPLE}} を両方含む）
	// When: 数値 placeholder を埋める
	got := embedManuscriptDraftLimits(constants.TextWriterBriefPrompt)

	// Then: 数値 placeholder は個別に消え、動的 placeholder は残る
	for _, ph := range []string{"{{TITLE_MIN}}", "{{PREFACE_MAX}}", "{{TOTAL_TARGET}}"} {
		if strings.Contains(got, ph) {
			t.Fatalf("数値 placeholder %s が残っている", ph)
		}
	}
	for _, ph := range []string{"{{SOURCES}}", "{{JSON_EXAMPLE}}"} {
		if !strings.Contains(got, ph) {
			t.Fatalf("動的 placeholder %s が消えている", ph)
		}
	}
}

func TestEmbedManuscriptDraftLimits_leavesNoNumericPlaceholder_whenTemplateListsAll(t *testing.T) {
	t.Parallel()

	// Given: manuscript_draft_limits の全数値 placeholder だけを列挙した template
	// （動的 placeholder を含まないので {{ の全消えで網羅を確認できる）
	numericPlaceholders := []string{
		"{{TITLE_MIN}}", "{{TITLE_MAX}}", "{{TITLE_TARGET}}",
		"{{INTRO_MIN}}", "{{INTRO_MAX}}", "{{INTRO_TARGET}}",
		"{{CLOSING_MIN}}", "{{CLOSING_MAX}}", "{{CLOSING_TARGET}}",
		"{{TOPIC_TITLE_MIN}}", "{{TOPIC_TITLE_MAX}}", "{{TOPIC_TITLE_TARGET}}",
		"{{PREFACE_MIN}}", "{{PREFACE_MAX}}", "{{PREFACE_TARGET}}",
		"{{DETAIL_MIN}}", "{{DETAIL_MAX}}", "{{DETAIL_TARGET}}",
		"{{TOPIC_COUNT_MIN}}", "{{TOPIC_COUNT_MAX}}", "{{TOPIC_COUNT_TARGET}}",
		"{{TOTAL_MIN}}", "{{TOTAL_MAX}}", "{{TOTAL_TARGET}}",
	}
	template := strings.Join(numericPlaceholders, " ")

	// When: 数値 placeholder を埋める
	got := embedManuscriptDraftLimits(template)

	// Then: 列挙した数値 placeholder が全て消えている
	if strings.Contains(got, "{{") || strings.Contains(got, "}}") {
		t.Fatalf("数値 placeholder が残っている: %q", got)
	}
}
