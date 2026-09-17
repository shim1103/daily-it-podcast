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

// writerOutputWithTotalRunes は合計対象 field の trim 後 rune 合計がちょうど n になる WriterOutput を返す。
// Intro へ日本語 rune「あ」を n 個並べ、他 field は空文字にする。Intro は合計対象、Title は対象外。
// utf8.RuneCountInString("あ") == 1 なので validateTotalChars が単純合計する total は n に一致する。
// field range・日本語含有・末尾句点は validateTotalChars が見ないため考慮しない。
func writerOutputWithTotalRunes(n int) models.WriterOutput {
	return models.WriterOutput{
		Intro: strings.Repeat("あ", n),
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

func TestValidateTotalChars_returnsNil_whenTotalEqualsMin(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMin ちょうどの WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMin)

	// When: 本番 topic 数で validateTotalChars を呼ぶ
	err := validateTotalChars(w, constants.DraftTopicCountTarget)

	// Then: error なし
	if err != nil {
		t.Fatalf("validateTotalChars: 下限ちょうどで予期しない error: %v", err)
	}
}

func TestValidateTotalChars_returnsNil_whenTotalEqualsMax(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMax ちょうどの WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMax)

	// When: 本番 topic 数で validateTotalChars を呼ぶ
	err := validateTotalChars(w, constants.DraftTopicCountTarget)

	// Then: error なし
	if err != nil {
		t.Fatalf("validateTotalChars: 上限ちょうどで予期しない error: %v", err)
	}
}

// --- 境界: total 範囲外 ---

func TestValidateTotalChars_returnsInvalidManuscriptDraft_whenTotalBelowMin(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMin - 1 の WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMin - 1)

	// When: 本番 topic 数で validateTotalChars を呼ぶ
	err := validateTotalChars(w, constants.DraftTopicCountTarget)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("validateTotalChars: error を期待したが nil")
	}
	assertTotalCharsRejected(t, err)
}

func TestValidateTotalChars_returnsInvalidManuscriptDraft_whenTotalAboveMax(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field の rune 合計が DraftTotalCharsMax + 1 の WriterOutput
	w := writerOutputWithTotalRunes(constants.DraftTotalCharsMax + 1)

	// When: 本番 topic 数で validateTotalChars を呼ぶ
	err := validateTotalChars(w, constants.DraftTopicCountTarget)

	// Then: invalid_manuscript_draft の Domain Error
	if err == nil {
		t.Fatalf("validateTotalChars: error を期待したが nil")
	}
	assertTotalCharsRejected(t, err)
}

// --- topicCount 引数化の反応確認 ---

// TestValidateTotalChars_usesTopicCountArgument_forRangeSelection は、
// validateTotalChars の range が expectedTopicCount 引数から導出され、
// constants.DraftTotalCharsMin/Max という DraftTopicCountTarget 固定値を
// 直接参照し続けているだけではないことを固定する。
func TestValidateTotalChars_usesTopicCountArgument_forRangeSelection(t *testing.T) {
	t.Parallel()

	const smallTopicCount = 1

	// Given: topic 数 1 前提の全体文字数下限ちょうどの WriterOutput
	// （DraftTotalCharsMin は topic 数 5 前提の値なので、topic 数 1 の下限とは異なる）
	smallMin := constants.TotalCharsMinFor(smallTopicCount)
	if smallMin >= constants.DraftTotalCharsMin {
		t.Fatalf("前提が崩れている: TotalCharsMinFor(%d)=%d は DraftTotalCharsMin=%d 未満であるべき", smallTopicCount, smallMin, constants.DraftTotalCharsMin)
	}
	w := writerOutputWithTotalRunes(smallMin)

	// When: topic 数 1 を渡して validateTotalChars を呼ぶ
	err := validateTotalChars(w, smallTopicCount)

	// Then: topic 数 1 の range では下限ちょうどなので error なし
	if err != nil {
		t.Fatalf("validateTotalChars: topicCount=%d の下限ちょうどで予期しない error: %v", smallTopicCount, err)
	}

	// When: 同じ WriterOutput を DraftTopicCountTarget（5）の range で検証する
	errWithTargetCount := validateTotalChars(w, constants.DraftTopicCountTarget)

	// Then: topic 数 5 の下限には届かないため invalid になる（range が引数依存で切り替わっている証拠）
	if errWithTargetCount == nil {
		t.Fatalf("validateTotalChars: topicCount=%d では下限未達のはずが error なし", constants.DraftTopicCountTarget)
	}
	assertTotalCharsRejected(t, errWithTargetCount)
}

// helper 健全性: writerOutputWithTotalRunes の rune 合計が引数に一致することを保証する。
// helper がずれると上の 4 case の境界前提が崩れるため独立に検算する。
func TestWriterOutputWithTotalRunes_sumsToArgument(t *testing.T) {
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
		// 合計対象は intro + closingSummary + Σ_topics(preface + detail)。title / topic.title は足さない。
		total := utf8.RuneCountInString(strings.TrimSpace(w.Intro)) +
			utf8.RuneCountInString(strings.TrimSpace(w.ClosingSummary))
		for _, tp := range w.Topics {
			total += utf8.RuneCountInString(strings.TrimSpace(tp.Preface))
			total += utf8.RuneCountInString(strings.TrimSpace(tp.Detail))
		}
		if total != n {
			t.Fatalf("writerOutputWithTotalRunes(%d): rune 合計 = %d, want %d", n, total, n)
		}
	}
}
