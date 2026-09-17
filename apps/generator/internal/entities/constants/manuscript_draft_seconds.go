package constants

// ManuscriptDraft の朗読 field 尺の正本（秒）。
// 文字数（manuscript_draft_limits.go の Draft*Len 系）は、ここの秒数に CharsPerSecond を
// 掛けた const 畳み込みで導出する。build 層は「秒」を知らず文字数だけを見る。
//
// openingGreeting / closingFarewell は TTS 前に定型文で付与するため、この尺・合計に含めない。
// title・topic.title は「朗読されない見出し」なので尺を持たない（manuscript_draft_limits.go で
// 秒非依存の文字数のみ定義）。
//
// margin 系定数（*MarginSec）だけを source of truth とし、Min/Max は target ± margin の
// const 畳み込みで導出する。topic 側の target・margin は DraftTopicSec / TopicMarginSec を
// 1 の位置に固定し、topic 数（DraftTopicCountTarget）を掛けた合計値から全体へ積み上げる。
// topic 数を変えても他の定数を壊さない（Σ_field margin > 全体 margin の契約は topic 数に
// 依存しない形にしてある）。整数演算の割り算は常に切り上げを想定する。
//
// Decision: docs/decisions/2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md

// CharsPerSecond は日本語 TTS のやや速めの発話速度（1 秒あたり文字数）。
const CharsPerSecond = 7

const (
	// DraftIntroTgtSec / DraftClosingTgtSec は intro・closingSummary の target 秒数。
	DraftIntroTgtSec   = 30
	DraftClosingTgtSec = 30

	IntroMarginSec   = 10
	ClosingMarginSec = 10

	DraftIntroMinSec = DraftIntroTgtSec - IntroMarginSec
	DraftIntroMaxSec = DraftIntroTgtSec + IntroMarginSec

	DraftClosingMinSec = DraftClosingTgtSec - ClosingMarginSec
	DraftClosingMaxSec = DraftClosingTgtSec + ClosingMarginSec
)

const (
	// DraftTopicSec は 1 topic あたりの target 秒数（preface + detail 合計）。
	DraftTopicSec = 108

	// TopicMarginSec は 1 topic（preface + detail 合計）あたりの target からの ± 幅（秒）。
	TopicMarginSec = 30

	// TopicPrefaceDetailRateNum / TopicPrefaceDetailRateDen は preface:detail の按分比
	// （1:4）のうち detail 側の分子・分母。target・margin 双方をこの比で按分する。
	TopicPrefaceDetailRateNum = 4
	TopicPrefaceDetailRateDen = 5

	// DraftTopicDetailTgtSec は DraftTopicSec を 1:4 (preface:detail) で按分した detail 側。
	// why: 切り上げを detail 側に寄せ、余りを preface 側に落とす（常に割り切れる保証をしない）。
	DraftTopicDetailTgtSec = (DraftTopicSec*TopicPrefaceDetailRateNum + TopicPrefaceDetailRateDen - 1) / TopicPrefaceDetailRateDen
	// DraftTopicPrefaceTgtSec は DraftTopicSec から DraftTopicDetailTgtSec を引いた残り。
	DraftTopicPrefaceTgtSec = DraftTopicSec - DraftTopicDetailTgtSec

	// TopicDetailMarginSec / TopicPrefaceMarginSec は TopicMarginSec を同じ 1:4 で按分した値。
	// why: target と同じ按分規則（detail 側切り上げ、preface 側は残り）で margin も配分する。
	TopicDetailMarginSec  = (TopicMarginSec*TopicPrefaceDetailRateNum + TopicPrefaceDetailRateDen - 1) / TopicPrefaceDetailRateDen
	TopicPrefaceMarginSec = TopicMarginSec - TopicDetailMarginSec

	DraftTopicPrefaceMinSec = DraftTopicPrefaceTgtSec - TopicPrefaceMarginSec
	DraftTopicPrefaceMaxSec = DraftTopicPrefaceTgtSec + TopicPrefaceMarginSec

	DraftTopicDetailMinSec = DraftTopicDetailTgtSec - TopicDetailMarginSec
	DraftTopicDetailMaxSec = DraftTopicDetailTgtSec + TopicDetailMarginSec
)

const (
	// TotalMarginRateNum / TotalMarginRateDen は「topic margin 合計のうち全体 margin へ
	// 繰り込む比率」（1:2）。値そのものは shim 指定。
	TotalMarginRateNum = 1
	TotalMarginRateDen = 2

	// DraftTotalTgtSec は挨拶を除いた朗読 field の合計 target（秒）。
	// intro + closing + topic 合計（DraftTopicSec × DraftTopicCountTarget）。
	// why: Go の const 式は関数呼び出しを許さないため、本番値は畳み込み式のまま const
	// で持つ。任意 topic 数向けの一般形は TotalTgtSecFor に用意し、両者が
	// DraftTopicCountTarget で一致することは contract test（manuscript_draft_seconds_contract_test.go）
	// が保証する。
	DraftTotalTgtSec = DraftIntroTgtSec + DraftClosingTgtSec + DraftTopicSec*DraftTopicCountTarget

	// TotalMarginSec は全体 target からの ± 幅（秒）。
	// why: topic 数ぶんの TopicMarginSec をそのまま全部足すと「各 field を min ギリギリに
	// 揃えても全体 min に届く」余裕が topic 数に比例して増え続け、圧力にならない
	// （shim 指摘）。rate で圧縮することで、topic 数を変えても
	// Σ(各 field margin) > TotalMarginSec の圧力が壊れない形にする。
	// why: 割り算は常に切り上げ（(a*num + den - 1) / den）を想定する。
	// why: DraftTotalTgtSec 同様、本番値は畳み込み式のまま const で持ち、一般形は
	// TotalMarginSecFor に用意する。
	TotalMarginSec = IntroMarginSec + ClosingMarginSec +
		(TopicMarginSec*DraftTopicCountTarget*TotalMarginRateNum+TotalMarginRateDen-1)/TotalMarginRateDen

	DraftTotalMinSec = DraftTotalTgtSec - TotalMarginSec
	DraftTotalMaxSec = DraftTotalTgtSec + TotalMarginSec
)

// TotalTgtSecFor は topicCount 件の topic を前提にした、挨拶を除く朗読 field 合計の
// target 秒数を返す。DraftTotalTgtSec の畳み込み式を任意 topic 数へ一般化した関数版。
//
// @require topicCount >= 0。
// @ensure 戻りは DraftIntroTgtSec + DraftClosingTgtSec + DraftTopicSec*topicCount。
// @ensure TotalTgtSecFor(DraftTopicCountTarget) == DraftTotalTgtSec。
func TotalTgtSecFor(topicCount int) int {
	return DraftIntroTgtSec + DraftClosingTgtSec + DraftTopicSec*topicCount
}

// TotalMarginSecFor は topicCount 件の topic を前提にした全体 margin（秒）を返す。
// TotalMarginSec の畳み込み式を任意 topic 数へ一般化した関数版。
//
// @require topicCount >= 0。
// @ensure 戻りは IntroMarginSec + ClosingMarginSec + ceil(TopicMarginSec*topicCount*TotalMarginRateNum / TotalMarginRateDen)。
// @ensure TotalMarginSecFor(DraftTopicCountTarget) == TotalMarginSec。
func TotalMarginSecFor(topicCount int) int {
	return IntroMarginSec + ClosingMarginSec +
		(TopicMarginSec*topicCount*TotalMarginRateNum+TotalMarginRateDen-1)/TotalMarginRateDen
}

// TotalMinSecFor / TotalMaxSecFor は topicCount 件の topic を前提にした全体尺の
// Min/Max（秒）を返す。DraftTotalMinSec / DraftTotalMaxSec の一般化。
//
// @require topicCount >= 0。
// @ensure TotalMinSecFor(topicCount) == TotalTgtSecFor(topicCount) - TotalMarginSecFor(topicCount)。
// @ensure TotalMaxSecFor(topicCount) == TotalTgtSecFor(topicCount) + TotalMarginSecFor(topicCount)。
func TotalMinSecFor(topicCount int) int {
	return TotalTgtSecFor(topicCount) - TotalMarginSecFor(topicCount)
}

func TotalMaxSecFor(topicCount int) int {
	return TotalTgtSecFor(topicCount) + TotalMarginSecFor(topicCount)
}
