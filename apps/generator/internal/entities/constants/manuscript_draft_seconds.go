package constants

// ManuscriptDraft の朗読 field 尺の正本（秒）。
// 文字数（manuscript_draft_limits.go の Draft*Len 系）は、ここの秒数に CharsPerSecond を
// 掛けた const 畳み込みで導出する。build 層は「秒」を知らず文字数だけを見る。
//
// openingGreeting / closingFarewell は TTS 前に定型文で付与するため、この尺・合計に含めない。
// title・topic.title は「朗読されない見出し」なので尺を持たない（manuscript_draft_limits.go で
// 秒非依存の文字数のみ定義）。
//
// Decision: docs/decisions/2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md

// CharsPerSecond は日本語 TTS のやや速めの発話速度（1 秒あたり文字数）。
const CharsPerSecond = 7

const (
	// intro（今日の episode の導入）。
	DraftIntroMinSec = 20
	DraftIntroTgtSec = 30
	DraftIntroMaxSec = 40

	// closingSummary（まとめ。挨拶で締めない）。
	DraftClosingMinSec = 20
	DraftClosingTgtSec = 30
	DraftClosingMaxSec = 40

	// 各 topic の preface（その topic への短い前置き）。
	// why: Prompt は「短い前置き」だが旧 Min=20s は短くない。System 実測で 99〜133 rune
	// （約 14〜19s）が繰り返し出たため、下限を 10s に合わせる。
	DraftTopicPrefaceMinSec = 10
	DraftTopicPrefaceTgtSec = 28
	DraftTopicPrefaceMaxSec = 36

	// 各 topic の detail（ソースに基づく説明本文）。
	// why: System 実測で detail 348 rune（下限 350）の惜しい不足が観測された（run 33308073574）。
	DraftTopicDetailMinSec = 48
	DraftTopicDetailTgtSec = 80
	DraftTopicDetailMaxSec = 110

	// episode 全体尺（挨拶を除く朗読 field の合計）。
	DraftTotalMinSec = 8 * 60  // 8 分
	DraftTotalTgtSec = 10 * 60 // 10 分
	DraftTotalMaxSec = 12 * 60 // 12 分
)
