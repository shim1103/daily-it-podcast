package build

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	domainerr "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// writerOutputWithTotalRunes は全朗読 field の trim 後 rune 合計がちょうど n になる WriterOutput を返す。
// Title へ日本語 rune「あ」を n 個並べ、他 field は空文字にする。utf8.RuneCountInString("あ") == 1 なので
// validateTotalChars が単純合計する total は n に一致する。field range・日本語含有・末尾句点は
// validateTotalChars が見ないため考慮しない。
func writerOutputWithTotalRunes(n int) models.WriterOutput {
	return models.WriterOutput{
		Title: strings.Repeat("あ", n),
	}
}

// assertTotalCharsRejected は err が invalid_manuscript_draft の Domain Error で cause 非空であることを確認する。
func assertTotalCharsRejected(t *testing.T, err error) {
	t.Helper()
	var de *domainerr.Error
	if !errors.As(err, &de) {
		t.Fatalf("validateTotalChars: *errors.Error へ型 assert 失敗: %T (%v)", err, err)
	}
	if de.Op != domainerr.OpInvalidManuscriptDraft {
		t.Fatalf("validateTotalChars: Op = %q, want %q", de.Op, domainerr.OpInvalidManuscriptDraft)
	}
	if de.Err == nil || strings.TrimSpace(de.Err.Error()) == "" {
		t.Fatalf("validateTotalChars: cause が空")
	}
}

// --- 正常系: total 範囲内ちょうど ---

func TestValidateTotalChars_total文字数が下限ちょうどなら受理する(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMin ちょうどの WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMin)

	// When: validateTotalChars を呼ぶ
	err := validateTotalChars(w)

	// Then: error なし
	if err != nil {
		t.Fatalf("validateTotalChars: 下限ちょうどで予期しない error: %v", err)
	}
}

func TestValidateTotalChars_total文字数が上限ちょうどなら受理する(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMax ちょうどの WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMax)

	// When: validateTotalChars を呼ぶ
	err := validateTotalChars(w)

	// Then: error なし
	if err != nil {
		t.Fatalf("validateTotalChars: 上限ちょうどで予期しない error: %v", err)
	}
}

// --- 境界: total 範囲外 ---

func TestValidateTotalChars_total文字数が下限未満ならreject(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMin - 1 の WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMin - 1)

	// When: validateTotalChars を呼ぶ
	err := validateTotalChars(w)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("validateTotalChars: error を期待したが nil")
	}
	assertTotalCharsRejected(t, err)
}

func TestValidateTotalChars_total文字数が上限超過ならreject(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMax + 1 の WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMax + 1)

	// When: validateTotalChars を呼ぶ
	err := validateTotalChars(w)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("validateTotalChars: error を期待したが nil")
	}
	assertTotalCharsRejected(t, err)
}

// helper 健全性: writerOutputWithTotalRunes の rune 合計が引数に一致することを保証する。
// helper がずれると上の 4 case の境界前提が崩れるため独立に検算する。
func TestWriterOutputWithTotalRunes_rune合計が引数に一致する(t *testing.T) {
	t.Parallel()

	// Given: 代表的な rune 数
	for _, n := range []int{
		constants.DraftTotalCharsMin - 1,
		constants.DraftTotalCharsMin,
		constants.DraftTotalCharsMax,
		constants.DraftTotalCharsMax + 1,
	} {
		// When: writerOutputWithTotalRunes で WriterOutput を組む
		w := writerOutputWithTotalRunes(n)

		// Then: validateTotalChars と同じ数え方の合計が n に一致する
		total := utf8.RuneCountInString(strings.TrimSpace(w.Title)) +
			utf8.RuneCountInString(strings.TrimSpace(w.Intro)) +
			utf8.RuneCountInString(strings.TrimSpace(w.ClosingSummary))
		for _, tp := range w.Topics {
			total += utf8.RuneCountInString(strings.TrimSpace(tp.Title))
			total += utf8.RuneCountInString(strings.TrimSpace(tp.Preface))
			total += utf8.RuneCountInString(strings.TrimSpace(tp.Detail))
		}
		if total != n {
			t.Fatalf("writerOutputWithTotalRunes(%d): rune 合計 = %d, want %d", n, total, n)
		}
	}
}
