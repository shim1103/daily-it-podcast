package build_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

func tunableSeedItems() []models.SourceItem {
	return []models.SourceItem{
		{
			SourceID:   "seed-1",
			OccurredAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Context:    "あるクラウド事業者がコンテナ基盤を更新した。オートスケール閾値を自動推定する。",
		},
	}
}

func TestComposeBriefWithTemplate_embedsGivenTemplate_whenCustomPromptProvided(t *testing.T) {
	t.Parallel()

	// Given: 数値 placeholder と {{SOURCES}} / {{JSON_EXAMPLE}} を含む任意 template
	template := "detail は {{DETAIL_MIN}}〜{{DETAIL_MAX}} 文字（目安 {{DETAIL_TARGET}}）。\n" +
		"# Source\n{{SOURCES}}\n# Example\n{{JSON_EXAMPLE}}\n独自マーカー_XYZ"

	// When: 任意 template で brief を組む
	got, err := build.ComposeBriefWithTemplate(tunableSeedItems(), template)

	// Then: 数値 placeholder と動的 placeholder がすべて埋まり、template 固有文字列が残る
	if err != nil {
		t.Fatalf("ComposeBriefWithTemplate: %v", err)
	}
	if strings.Contains(got, "{{") || strings.Contains(got, "}}") {
		t.Fatalf("placeholder が残っている: %q", got)
	}
	if !strings.Contains(got, "独自マーカー_XYZ") {
		t.Fatalf("template 固有文字列が消えている: %q", got)
	}
	if !strings.Contains(got, "seed-1") {
		t.Fatalf("{{SOURCES}} が埋まっていない: %q", got)
	}
	if !strings.Contains(got, strconv.Itoa(constants.DraftTopicDetailMinLen)) {
		t.Fatalf("{{DETAIL_MIN}} が limits 値で埋まっていない: %q", got)
	}
}

func TestComposeBriefWithTemplate_matchesComposeBrief_whenDefaultTemplateGiven(t *testing.T) {
	t.Parallel()

	// Given: 同一 items
	items := tunableSeedItems()

	// When: 既定 template を明示的に渡した場合と ComposeBrief の出力を比べる
	viaTemplate, err := build.ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt)
	if err != nil {
		t.Fatalf("ComposeBriefWithTemplate(default): %v", err)
	}
	viaDefault, err := build.ComposeBrief(items)
	if err != nil {
		t.Fatalf("ComposeBrief: %v", err)
	}

	// Then: 出力は完全一致（委譲による不変）
	if viaTemplate != viaDefault {
		t.Fatalf("出力不一致:\n--- template ---\n%s\n--- default ---\n%s", viaTemplate, viaDefault)
	}
}
