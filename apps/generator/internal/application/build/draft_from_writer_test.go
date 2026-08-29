package build_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerr "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

// jaRunes は日本語 rune（ひらがな）を n 個並べた文字列を返す。
func jaRunes(n int) string {
	return strings.Repeat("あ", n)
}

// jaField は日本語 n rune + 末尾句点（合計 n+1 rune）の朗読 field を返す。
func jaField(n int) string {
	return jaRunes(n) + string(constants.DraftSentenceSuffixRune)
}

// topicJSON は 3 field すべて runes rune + 句点 の topic を JSON 断片で返す。
func topicJSON(runes int) string {
	f := jaField(runes)
	return `{"title":"` + f + `","preface":"` + f + `","detail":"` + f + `"}`
}

// wireOverride は buildWireJSON の default 生成値を field 単位で差し替える。
// nil の field は default を使い、非 nil の field はその raw string 値を JSON quote 内へ埋め込む。
// これにより reject fixture を「意図した field を明示的に不正化」した形で直接組み立てられる。
type wireOverride struct {
	title   *string // "title" の値（末尾句点なし・ASCII のみ・空白のみ・rune 数逸脱などを注入する）
	closing *string // "closingSummary" の値（rune 数 range 逸脱などを注入する）
}

// buildWireJSON は指定 topic 数の wire JSON を組み立てる。
// title / intro / closing / 各 topic field はいずれも rune 数を range 上限付近（max-1 rune）に取る。
func buildWireJSON(topicCount int) string {
	return buildWireJSONWith(topicCount, wireOverride{})
}

// buildWireJSONWith は buildWireJSON に field 上書きを加えて wire JSON を組み立てる。
// title / intro / closing / 各 topic field はいずれも default では range 内に収める。
func buildWireJSONWith(topicCount int, ov wireOverride) string {
	title := jaField(constants.DraftTitleMaxLen - 1)
	intro := jaField(constants.DraftIntroMaxLen - 1)
	closing := jaField(constants.DraftClosingMaxLen - 1)

	if ov.title != nil {
		title = *ov.title
	}
	if ov.closing != nil {
		closing = *ov.closing
	}

	topics := make([]string, 0, topicCount)
	for i := 0; i < topicCount; i++ {
		topics = append(topics, topicJSON(constants.DraftTopicTitleMaxLen-1))
	}
	return `{"title":"` + title + `","intro":"` + intro +
		`","topics":[` + strings.Join(topics, ",") +
		`],"closingSummary":"` + closing + `"}`
}

// strPtr は string literal を *string へ変換する（wireOverride 用）。
func strPtr(s string) *string {
	return &s
}

// validWire は Acceptance を満たす標準 wire を返す（全 field range・total range とも満たす）。
// fixture が実際に valid であることは TestBuildWireJSON_最大topic数のfixtureはvalidである が保証する。
func validWire() string {
	return buildWireJSON(constants.DraftTopicCountMax)
}

// assertInvalidDraft は err が invalid_manuscript_draft の Domain Error で cause 非空であることを確認する。
func assertInvalidDraft(t *testing.T, err error) {
	t.Helper()
	var de *domainerr.Error
	if !errors.As(err, &de) {
		t.Fatalf("ManuscriptDraftFromWriterOutput: *errors.Error へ型 assert 失敗: %T (%v)", err, err)
	}
	if de.Op != domainerr.OpInvalidManuscriptDraft {
		t.Fatalf("ManuscriptDraftFromWriterOutput: Op = %q, want %q", de.Op, domainerr.OpInvalidManuscriptDraft)
	}
	if de.Err == nil || strings.TrimSpace(de.Err.Error()) == "" {
		t.Fatalf("ManuscriptDraftFromWriterOutput: cause が空")
	}
}

// --- fixture 自己検証 ---

func TestBuildWireJSON_最大topic数のfixtureはvalidである(t *testing.T) {
	t.Parallel()

	// Given: buildWireJSON が生成する最大 topic 数の wire
	raw := buildWireJSON(constants.DraftTopicCountMax)

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: 全 field range・topic 数 range・total range を満たし error なし
	if err != nil {
		t.Fatalf("buildWireJSON(DraftTopicCountMax): fixture が valid でない: %v", err)
	}
}

// --- 正常系 ---

func TestManuscriptDraftFromWriterOutput_validWireをDraftへ変換する(t *testing.T) {
	t.Parallel()

	// Given: limits を満たす JSON wire
	raw := validWire()

	// When: parse する
	got, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: error なし
	if err != nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: 予期しない error: %v", err)
	}
	// Then: Title / Intro / ClosingSummary が転記される
	if got.Title == "" || got.Intro == "" || got.ClosingSummary == "" {
		t.Fatalf("ManuscriptDraftFromWriterOutput: field 転記漏れ: %+v", got)
	}
	// Then: topic 数が一致する
	if len(got.Topics) != constants.DraftTopicCountMax {
		t.Fatalf("ManuscriptDraftFromWriterOutput: topic 数 = %d, want %d", len(got.Topics), constants.DraftTopicCountMax)
	}
	// Then: topic field も転記される
	for i, tp := range got.Topics {
		if tp.Title == "" || tp.Preface == "" || tp.Detail == "" {
			t.Fatalf("ManuscriptDraftFromWriterOutput: topic[%d] 転記漏れ: %+v", i, tp)
		}
	}
}

func TestManuscriptDraftFromWriterOutput_codeフェンス付きwireも受理する(t *testing.T) {
	t.Parallel()

	// Given: ```json ... ``` で括られた valid wire
	raw := "```json\n" + validWire() + "\n```"

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: error なし
	if err != nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: code fence strip 失敗: %v", err)
	}
}

// --- 異常系 ---

func TestManuscriptDraftFromWriterOutput_JSONとして不正ならreject(t *testing.T) {
	t.Parallel()

	// Given: 途中で切れて object を閉じない wire
	raw := `{"title": "あ。", "intro":`

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_wireが空文字ならreject(t *testing.T) {
	t.Parallel()

	// Given: 空白のみで trim 後に空になる wire
	raw := "   "

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_朗読fieldの末尾が句点でないならreject(t *testing.T) {
	t.Parallel()

	// Given: title を日本語 rune のみ（末尾句点なし）へ上書きした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(jaRunes(constants.DraftTitleTargetLen)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_朗読fieldに日本語がないならreject(t *testing.T) {
	t.Parallel()

	// Given: title を ASCII 英字のみ + 句点 へ上書きした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(strings.Repeat("a", constants.DraftTitleTargetLen-1) + string(constants.DraftSentenceSuffixRune)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_朗読fieldが空白のみならreject(t *testing.T) {
	t.Parallel()

	// Given: title を空白のみへ上書きした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr("    "),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

// --- 境界: topic 数 ---

func TestManuscriptDraftFromWriterOutput_topic数が下限未満ならreject(t *testing.T) {
	t.Parallel()

	// Given: topic 数を下限 - 1 にした wire
	raw := buildWireJSON(constants.DraftTopicCountMin - 1)

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_topic数が上限超過ならreject(t *testing.T) {
	t.Parallel()

	// Given: topic 数を上限 + 1 にした wire
	raw := buildWireJSON(constants.DraftTopicCountMax + 1)

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

// --- 境界: title rune 数 ---

func TestManuscriptDraftFromWriterOutput_titleのrune数が下限未満ならreject(t *testing.T) {
	t.Parallel()

	// Given: title の rune 数を下限 - 1 にした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(jaField(constants.DraftTitleMinLen - 2)), // rune 数 = Min-1
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_titleのrune数が上限超過ならreject(t *testing.T) {
	t.Parallel()

	// Given: title の rune 数を上限 + 1 にした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(jaField(constants.DraftTitleMaxLen)), // rune 数 = Max+1
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

// --- 境界: closingSummary rune 数 ---

func TestManuscriptDraftFromWriterOutput_closingSummaryのrune数が下限未満ならreject(t *testing.T) {
	t.Parallel()

	// Given: closingSummary の rune 数を下限 - 1 にした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		closing: strPtr(jaField(constants.DraftClosingMinLen - 2)), // rune 数 = Min-1
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_closingSummaryのrune数が上限超過ならreject(t *testing.T) {
	t.Parallel()

	// Given: closingSummary の rune 数を上限 + 1 にした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		closing: strPtr(jaField(constants.DraftClosingMaxLen)), // rune 数 = Max+1
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}
