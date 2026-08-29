package constants

import "testing"

// manuscript_draft_limits.go / manuscript_draft_seconds.go の定数群が満たすべき
// invariant（Design by Contract の不変条件）を固定する contract test。
//
// 検証するのは「定数どうしの数学的関係」のみ。個々の定数値の現実妥当性
// （何文字が何分か等）は、その値を消費する validateTotalChars の unit test が担う。
// 畳み込み式そのものの再掲（DraftIntroMinLen == DraftIntroMinSec*CharsPerSecond 等）は
// 自己参照 assert で検出力を持たないため書かない。

// narrationFieldSumChars は挨拶を除いた朗読 field の合計文字数を返す。
// 合計対象は intro + closingSummary + topicCount 個の (preface + detail)。
// title・topic.title は「朗読されない見出し」なので合計に入れない。
func narrationFieldSumChars(introChars, closingChars, prefaceChars, detailChars, topicCount int) int {
	return introChars + closingChars + topicCount*(prefaceChars+detailChars)
}

func TestManuscriptDraftLimits_minConfigSumWithinTotalMax(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field を下限・topic 数も下限にした最小構成
	minSum := narrationFieldSumChars(
		DraftIntroMinLen, DraftClosingMinLen,
		DraftTopicPrefaceMinLen, DraftTopicDetailMinLen,
		DraftTopicCountMin,
	)

	// Then: 最小構成でも全体上限以下に収まる（収まらないと valid な draft を作れない）
	if minSum > DraftTotalCharsMax {
		t.Fatalf("最小構成の合計 %d が DraftTotalCharsMax %d を超える。field 下限か topic 数下限か total 上限が不整合", minSum, DraftTotalCharsMax)
	}
}

func TestManuscriptDraftLimits_maxConfigSumReachesTotalMin(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field を上限・topic 数も上限にした最大構成
	maxSum := narrationFieldSumChars(
		DraftIntroMaxLen, DraftClosingMaxLen,
		DraftTopicPrefaceMaxLen, DraftTopicDetailMaxLen,
		DraftTopicCountMax,
	)

	// Then: 最大構成で全体下限以上に届く（届かないと valid な draft を作れない）
	if maxSum < DraftTotalCharsMin {
		t.Fatalf("最大構成の合計 %d が DraftTotalCharsMin %d に届かない。field 上限か topic 数上限か total 下限が不整合", maxSum, DraftTotalCharsMin)
	}
}

func TestManuscriptDraftLimits_targetConfigSumEqualsTotalTarget(t *testing.T) {
	t.Parallel()

	// Given: 全朗読 field を target・topic 数も target にした構成
	tgtSum := narrationFieldSumChars(
		DraftIntroTarget, DraftClosingTarget,
		DraftTopicPrefaceTarget, DraftTopicDetailTarget,
		DraftTopicCountTarget,
	)

	// Then: 全体 target と一致する（target どうしの内部整合。ズレると prompt 誘導が矛盾する）
	if tgtSum != DraftTotalCharsTarget {
		t.Fatalf("target 構成の合計 %d が DraftTotalCharsTarget %d と一致しない", tgtSum, DraftTotalCharsTarget)
	}
}

func TestManuscriptDraftLimits_totalCharsBoundsAscend(t *testing.T) {
	t.Parallel()

	// Then: Min < Target < Max（typo や逆転の検出）
	if !(DraftTotalCharsMin < DraftTotalCharsTarget && DraftTotalCharsTarget < DraftTotalCharsMax) {
		t.Fatalf("全体文字数の順序が壊れている: Min=%d Target=%d Max=%d", DraftTotalCharsMin, DraftTotalCharsTarget, DraftTotalCharsMax)
	}
}

func TestManuscriptDraftLimits_topicCountBoundsAscend(t *testing.T) {
	t.Parallel()

	// Then: Min < Target < Max
	if !(DraftTopicCountMin < DraftTopicCountTarget && DraftTopicCountTarget < DraftTopicCountMax) {
		t.Fatalf("topic 数の順序が壊れている: Min=%d Target=%d Max=%d", DraftTopicCountMin, DraftTopicCountTarget, DraftTopicCountMax)
	}
}

func TestManuscriptDraftLimits_narrationFieldBoundsAscend(t *testing.T) {
	t.Parallel()

	// Given: 昇順であるべき朗読 field の (名前, Min, Target, Max)
	fields := []struct {
		name             string
		min, target, max int
	}{
		{"intro", DraftIntroMinLen, DraftIntroTarget, DraftIntroMaxLen},
		{"closingSummary", DraftClosingMinLen, DraftClosingTarget, DraftClosingMaxLen},
		{"topic.preface", DraftTopicPrefaceMinLen, DraftTopicPrefaceTarget, DraftTopicPrefaceMaxLen},
		{"topic.detail", DraftTopicDetailMinLen, DraftTopicDetailTarget, DraftTopicDetailMaxLen},
	}

	// Then: 各 field で Min <= Target <= Max
	for _, f := range fields {
		if !(f.min <= f.target && f.target <= f.max) {
			t.Fatalf("%s の順序が壊れている: Min=%d Target=%d Max=%d", f.name, f.min, f.target, f.max)
		}
	}
}

func TestManuscriptDraftLimits_headingFieldBoundsAscend(t *testing.T) {
	t.Parallel()

	// Given: 秒非依存の見出し field（合計対象外だが単体 range は持つ）
	fields := []struct {
		name             string
		min, target, max int
	}{
		{"title", DraftTitleMinLen, DraftTitleTargetLen, DraftTitleMaxLen},
		{"topic.title", DraftTopicTitleMinLen, DraftTopicTitleTarget, DraftTopicTitleMaxLen},
	}

	// Then: 各 field で Min <= Target <= Max
	for _, f := range fields {
		if !(f.min <= f.target && f.target <= f.max) {
			t.Fatalf("%s の順序が壊れている: Min=%d Target=%d Max=%d", f.name, f.min, f.target, f.max)
		}
	}
}
