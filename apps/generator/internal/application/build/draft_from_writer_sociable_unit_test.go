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
// 朗読 field（intro / closingSummary / topic.preface / topic.detail）は末尾句点を課されるため、
// 検証を通す default fixture はこの helper で作る。
func jaField(n int) string {
	return jaRunes(n) + string(constants.DraftSentenceSuffixRune)
}

// topicJSON は title を titleRunes rune の見出し（末尾句点なし）、
// preface / detail をそれぞれ prefaceRunes / detailRunes rune + 句点 の朗読 field にした
// topic を JSON 断片で返す。各 field の range は呼び出し側が対応する定数から個別に渡す。
func topicJSON(titleRunes, prefaceRunes, detailRunes int) string {
	return `{"title":"` + jaRunes(titleRunes) +
		`","preface":"` + jaField(prefaceRunes) +
		`","detail":"` + jaField(detailRunes) + `"}`
}

// wireOverride は buildWireJSON の default 生成値を field 単位で差し替える。
// nil の field は default を使い、非 nil の field はその raw string 値を JSON quote 内へ埋め込む。
// これにより reject fixture を「意図した field を明示的に不正化」した形で直接組み立てられる。
type wireOverride struct {
	title   *string // "title"（見出し。空白のみ・ASCII のみ・rune 数逸脱などを注入する）
	intro   *string // "intro"（朗読 field。末尾句点なし・ASCII のみ・空白のみなどを注入する）
	closing *string // "closingSummary"（朗読 field。rune 数 range 逸脱などを注入する）
}

// buildWireJSON は指定 topic 数の wire JSON を組み立てる。
// default の各 field 長は下で定義する buildWireJSONWith の方針に従う。
func buildWireJSON(topicCount int) string {
	return buildWireJSONWith(topicCount, wireOverride{})
}

// buildWireJSONWith は buildWireJSON に field 上書きを加えて wire JSON を組み立てる。
//
// default の field 長:
//   - title            見出し max-1 rune（range 内）
//   - intro / closing   朗読 field max-1 rune（range 内。境界付近を突く）
//   - topic.title       見出し max-1 rune（range 内）
//   - topic.preface     朗読 field min rune（range 内）
//   - topic.detail      朗読 field min rune（range 内）
//
// topic 数 max のとき合計対象は intro + closing + Σ_topics(preface + detail)。
// topic.preface / topic.detail を min に取るのは、topic 数 max × field max だと合計が
// DraftTotalCharsMax を超えるため。合計が range 内に収まることは
// TestBuildWireJSON_producesValidWire_whenTopicCountIsMax が保証する。
func buildWireJSONWith(topicCount int, ov wireOverride) string {
	title := jaRunes(constants.DraftTitleMaxLen - 1)
	intro := jaField(constants.DraftIntroMaxLen - 1)
	closing := jaField(constants.DraftClosingMaxLen - 1)

	if ov.title != nil {
		title = *ov.title
	}
	if ov.intro != nil {
		intro = *ov.intro
	}
	if ov.closing != nil {
		closing = *ov.closing
	}

	topics := make([]string, 0, topicCount)
	for i := 0; i < topicCount; i++ {
		topics = append(topics, topicJSON(
			constants.DraftTopicTitleMaxLen-1,
			constants.DraftTopicPrefaceMinLen,
			constants.DraftTopicDetailMinLen,
		))
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
// fixture が実際に valid であることは TestBuildWireJSON_producesValidWire_whenTopicCountIsMax が保証する。
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

func TestBuildWireJSON_producesValidWire_whenTopicCountIsMax(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_returnsDraft_whenWireIsValid(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_acceptsWire_whenWrappedInCodeFence(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_acceptsWire_whenProsePrecedesJSONObject(t *testing.T) {
	t.Parallel()

	// Given: Cursor 実測どおり先頭に日本語散文があり、その後ろに valid wire
	raw := "了解しました。以下が原稿です。\n" + validWire()

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: 先頭散文を落として wire として受理する
	if err != nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: 先頭散文付き wire の受理失敗: %v", err)
	}
}

// --- 異常系 ---

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenJSONIsMalformed(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenWireIsBlank(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenNarrationFieldLacksSentenceSuffix(t *testing.T) {
	t.Parallel()

	// Given: 朗読 field の intro を日本語 rune のみ（末尾句点なし）へ上書きした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		intro: strPtr(jaRunes(constants.DraftIntroTarget)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenNarrationFieldHasNoJapanese(t *testing.T) {
	t.Parallel()

	// Given: 朗読 field の intro を ASCII 英字のみ + 句点 へ上書きした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		intro: strPtr(strings.Repeat("a", constants.DraftIntroTarget-1) + string(constants.DraftSentenceSuffixRune)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenNarrationFieldIsWhitespaceOnly(t *testing.T) {
	t.Parallel()

	// Given: 朗読 field の intro を空白のみへ上書きした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		intro: strPtr("    "),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenHeadingFieldIsWhitespaceOnly(t *testing.T) {
	t.Parallel()

	// Given: 見出し field の title を空白のみへ上書きした wire（見出しも非空必須）
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenHeadingFieldHasNoJapanese(t *testing.T) {
	t.Parallel()

	// Given: 見出し field の title を ASCII 英字のみへ上書きした wire（見出しも日本語必須）
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(strings.Repeat("a", constants.DraftTitleTargetLen)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_acceptsHeadingField_whenItLacksSentenceSuffix(t *testing.T) {
	t.Parallel()

	// Given: 見出し field の title を日本語 rune のみ（末尾句点なし・range 内）にした wire
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(jaRunes(constants.DraftTitleTargetLen)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: 見出しは末尾句点を課されないため error なし
	if err != nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: 見出しの末尾句点なしで予期しない error: %v", err)
	}
}

// --- 境界: topic 数 ---

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenTopicCountBelowMin(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenTopicCountAboveMax(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenTitleRuneCountBelowMin(t *testing.T) {
	t.Parallel()

	// Given: title の rune 数を下限 - 1 にした wire（見出しなので句点なし日本語）
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(jaRunes(constants.DraftTitleMinLen - 1)),
	})

	// When: parse する
	_, err := build.ManuscriptDraftFromWriterOutput(raw)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("ManuscriptDraftFromWriterOutput: error を期待したが nil")
	}
	assertInvalidDraft(t, err)
}

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenTitleRuneCountAboveMax(t *testing.T) {
	t.Parallel()

	// Given: title の rune 数を上限 + 1 にした wire（見出しなので句点なし日本語）
	raw := buildWireJSONWith(constants.DraftTopicCountMax, wireOverride{
		title: strPtr(jaRunes(constants.DraftTitleMaxLen + 1)),
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenClosingSummaryRuneCountBelowMin(t *testing.T) {
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

func TestManuscriptDraftFromWriterOutput_returnsInvalidManuscriptDraft_whenClosingSummaryRuneCountAboveMax(t *testing.T) {
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
