package constants

import "testing"

// manuscript_draft_seconds.go / manuscript_draft_limits.go が新設した *For 関数群
// （TotalTgtSecFor 等）が、DraftTopicCountTarget を渡したとき既存 const 畳み込みと
// 同じ値を返すことを固定する contract test（Design by Contract の postcondition）。
//
// Go の const 式は関数呼び出しを許さないため、本番値は畳み込み式のまま const で持ち、
// 任意 topic 数向けの一般形は別関数として新設した（docs/decisions/2026-09-16T19-53-19-
// feature-generator-textwriter-adapter-fallback.md §1-4）。両者の乖離を防ぐのが本 test の役割。

func TestTotalTgtSecFor_matchesDraftTotalTgtSec_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalTgtSecFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != DraftTotalTgtSec {
		t.Fatalf("TotalTgtSecFor(%d) = %d, want %d", DraftTopicCountTarget, got, DraftTotalTgtSec)
	}
}

func TestTotalMarginSecFor_matchesTotalMarginSec_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalMarginSecFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != TotalMarginSec {
		t.Fatalf("TotalMarginSecFor(%d) = %d, want %d", DraftTopicCountTarget, got, TotalMarginSec)
	}
}

func TestTotalMinSecFor_matchesDraftTotalMinSec_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalMinSecFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != DraftTotalMinSec {
		t.Fatalf("TotalMinSecFor(%d) = %d, want %d", DraftTopicCountTarget, got, DraftTotalMinSec)
	}
}

func TestTotalMaxSecFor_matchesDraftTotalMaxSec_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalMaxSecFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != DraftTotalMaxSec {
		t.Fatalf("TotalMaxSecFor(%d) = %d, want %d", DraftTopicCountTarget, got, DraftTotalMaxSec)
	}
}

func TestTotalCharsMinFor_matchesDraftTotalCharsMin_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalCharsMinFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != DraftTotalCharsMin {
		t.Fatalf("TotalCharsMinFor(%d) = %d, want %d", DraftTopicCountTarget, got, DraftTotalCharsMin)
	}
}

func TestTotalCharsTargetFor_matchesDraftTotalCharsTarget_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalCharsTargetFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != DraftTotalCharsTarget {
		t.Fatalf("TotalCharsTargetFor(%d) = %d, want %d", DraftTopicCountTarget, got, DraftTotalCharsTarget)
	}
}

func TestTotalCharsMaxFor_matchesDraftTotalCharsMax_whenGivenDraftTopicCountTarget(t *testing.T) {
	t.Parallel()

	// When: 本番 topic 数を渡す
	got := TotalCharsMaxFor(DraftTopicCountTarget)

	// Then: 既存 const 畳み込みと一致する
	if got != DraftTotalCharsMax {
		t.Fatalf("TotalCharsMaxFor(%d) = %d, want %d", DraftTopicCountTarget, got, DraftTotalCharsMax)
	}
}

// TestTotalTgtSecFor_variesWithTopicCount は関数版が実際に topicCount へ反応することを固定する。
// 上の一致 test だけでは「常に DraftTotalTgtSec を返す定数関数」になっていても検出できない。
func TestTotalTgtSecFor_variesWithTopicCount(t *testing.T) {
	t.Parallel()

	// Given/When: topic 数 1 と DraftTopicCountTarget それぞれの target 秒数
	got1 := TotalTgtSecFor(1)
	gotTarget := TotalTgtSecFor(DraftTopicCountTarget)

	// Then: topic 数が異なれば値も異なる（DraftTopicSec 分だけ差が出る）
	if got1 == gotTarget {
		t.Fatalf("TotalTgtSecFor(1) と TotalTgtSecFor(%d) が同値 %d: topicCount に反応していない", DraftTopicCountTarget, got1)
	}
	if want := DraftTopicSec * (DraftTopicCountTarget - 1); gotTarget-got1 != want {
		t.Fatalf("TotalTgtSecFor の差分 = %d, want %d", gotTarget-got1, want)
	}
}
