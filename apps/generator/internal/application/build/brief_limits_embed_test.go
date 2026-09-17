package build

import (
	"errors"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerr "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

func TestEmbedManuscriptDraftLimits_replacesNumericPlaceholdersButKeepsDynamicOnes(t *testing.T) {
	t.Parallel()

	// Given: 本番で実際に渡る TextWriterBriefPrompt（数値 placeholder と
	// 動的 placeholder {{SOURCES}} / {{JSON_EXAMPLE}} を両方含む）
	// When: 本番 topic 数で数値 placeholder を埋める
	got := embedManuscriptDraftLimits(constants.TextWriterBriefPrompt, constants.DraftTopicCountTarget)

	// Then: 数値 placeholder は個別に消え、動的 placeholder は残る
	for _, ph := range []string{"{{TITLE_MIN}}", "{{PREFACE_MAX}}", "{{TOTAL_TARGET}}", "{{TOTAL_MINUTES_MIN}}", "{{TOTAL_MINUTES_MAX}}"} {
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
		"{{TOPIC_COUNT_TARGET}}",
		"{{TOTAL_MIN}}", "{{TOTAL_MAX}}", "{{TOTAL_TARGET}}",
		"{{TOTAL_MINUTES_MIN}}", "{{TOTAL_MINUTES_MAX}}",
	}
	template := strings.Join(numericPlaceholders, " ")

	// When: 本番 topic 数で数値 placeholder を埋める
	got := embedManuscriptDraftLimits(template, constants.DraftTopicCountTarget)

	// Then: 列挙した数値 placeholder が全て消えている
	if strings.Contains(got, "{{") || strings.Contains(got, "}}") {
		t.Fatalf("数値 placeholder が残っている: %q", got)
	}
}

// TestLoadWriterOutputExampleJSON_returnsRaw_whenValid は embed 済み example が
// loadWriterOutputExampleJSON を通り ManuscriptDraftFromWriterOutput も通ることを固定する。
func TestLoadWriterOutputExampleJSON_returnsRaw_whenValid(t *testing.T) {
	t.Parallel()

	// Given: embed 済みの WriterOutput JSON 平文（topic 数は固定 DraftTopicCountTarget 件）
	raw := strings.TrimSpace(writerOutputExampleJSON)

	// When: 読込と正当性検査をする
	got, err := loadWriterOutputExampleJSON(raw, constants.DraftTopicCountTarget)

	// Then: error なしで同一 JSON が返り、ManuscriptDraftFromWriterOutput も通る
	if err != nil {
		t.Fatalf("loadWriterOutputExampleJSON: %v", err)
	}
	if got != raw {
		t.Fatal("戻り JSON が入力と異なる")
	}
	if _, err := ManuscriptDraftFromWriterOutput(got, constants.DraftTopicCountTarget); err != nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: %v", err)
	}
}

// TestLoadWriterOutputExampleJSON_returnsValidationErrorAsIs_whenInvalid は
// invalid JSON のとき ManuscriptDraftFromWriterOutput の error を wrap せず返すことを固定する。
func TestLoadWriterOutputExampleJSON_returnsValidationErrorAsIs_whenInvalid(t *testing.T) {
	t.Parallel()

	// Given: draft 検証に落ちる短い JSON と、同一 raw を直接検証したときの error
	raw := `{"title":"短すぎる題","intro":"短い。","topics":[],"closingSummary":"短い。"}`
	wantErr := mustManuscriptDraftErrForExampleLoad(t, raw)

	// When: 読込検査する
	_, gotErr := loadWriterOutputExampleJSON(raw, constants.DraftTopicCountTarget)

	// Then: ManuscriptDraftFromWriterOutput と同じ Domain Error がそのまま返る
	if gotErr == nil {
		t.Fatal("error を期待したが nil")
	}
	assertSameDomainErrForExampleLoad(t, gotErr, wantErr)
}

func mustManuscriptDraftErrForExampleLoad(t *testing.T, raw string) error {
	t.Helper()
	_, err := ManuscriptDraftFromWriterOutput(raw, constants.DraftTopicCountTarget)
	if err == nil {
		t.Fatal("ManuscriptDraftFromWriterOutput が error を返さなかった")
	}
	return err
}

func assertSameDomainErrForExampleLoad(t *testing.T, got, want error) {
	t.Helper()
	var gotDE, wantDE *domainerr.Error
	if !errors.As(got, &gotDE) {
		t.Fatalf("got が Domain Error ではない: %T (%v)", got, got)
	}
	if !errors.As(want, &wantDE) {
		t.Fatalf("want が Domain Error ではない: %T (%v)", want, want)
	}
	if gotDE.Op != wantDE.Op {
		t.Fatalf("Op = %q, want %q", gotDE.Op, wantDE.Op)
	}
	if got.Error() != want.Error() {
		t.Fatalf("error 文言が変わっている\ngot:  %s\nwant: %s", got.Error(), want.Error())
	}
}
