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
	// why(tgt): 全体 target を round な 16 分（下記）に固定するため、topic 構成の端数を
	// intro / closingSummary の target で吸収する。8×(30+82)=896s の残り 64s を 32s ずつ分ける。
	DraftIntroMinSec = 20
	DraftIntroTgtSec = 32
	DraftIntroMaxSec = 40

	// closingSummary（まとめ。挨拶で締めない）。
	// why(tgt): intro と同じ理由で 30→32s。全体 target 16 分の端数吸収。
	DraftClosingMinSec = 20
	DraftClosingTgtSec = 32
	DraftClosingMaxSec = 40

	// 各 topic の preface（その topic への短い前置き）。
	// why: Prompt は「短い前置き」だが旧 Min=20s は短くない。System 実測で 99〜133 rune
	// （約 14〜19s）が繰り返し出たため、下限を 10s に合わせる。
	// why(tgt/max): topic 数 target が 5→8 へ増えたぶん 1 topic の負荷を軽くしつつ、
	// 全体 target 尺（下記）へ寄せるため tgt を 28→30s、max を 36→38s へ微調整する。
	DraftTopicPrefaceMinSec = 10
	DraftTopicPrefaceTgtSec = 30
	DraftTopicPrefaceMaxSec = 38

	// 各 topic の detail（ソースに基づく説明本文）。
	// why: System 実測で detail 348 rune（下限 350）の惜しい不足が観測された（run 33308073574）。
	// why(tgt/max): topic 数増に合わせ preface と同じ比率を保ったまま tgt を 80→82s、
	// max を 110→112s へ微調整する（preface+detail の target 合計 = 112s/topic）。
	DraftTopicDetailMinSec = 48
	DraftTopicDetailTgtSec = 82
	DraftTopicDetailMaxSec = 112

	// episode 全体尺（挨拶を除く朗読 field の合計）。
	// why: 1 episode の topic 数を増やしたぶん全体尺も伸ばす。Min/Max は「およそ 14〜18 分」
	// の意図をそのまま丸めた独立窓。Tgt は round な 16 分に固定し、contract test の
	// 「target 構成合計 = 全体 target」は intro/closing の target（各 32s）で端数を吸収して満たす
	// （32 + 32 + 8×(30+82) = 960s）。
	DraftTotalMinSec = 14 * 60 // 14 分
	DraftTotalTgtSec = 16 * 60 // 16 分
	DraftTotalMaxSec = 18 * 60 // 18 分
)
