package constants

// ManuscriptDraft / TextWriter 出力の parse・validation 用文字数定数（Domain Rule 正本）。
// 朗読 field（intro / closingSummary / topic.preface / topic.detail / 全体）の文字数は
// manuscript_draft_seconds.go の秒数 × CharsPerSecond の const 畳み込みで導出する。
// title / topic.title は朗読されない見出しなので秒非依存の文字数を直接定義する。
// embed 専用 Prompt 文言は text_writer_brief_prompt.go。数値 placeholder は ComposeBrief が本定数で埋める。
// Decision: docs/decisions/2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md

const (
	// title（見出し。朗読されず全体文字数にも数えない）。target ± margin の畳み込み。
	DraftTitleTargetLen = 40
	TitleMarginLen      = 10
	DraftTitleMinLen    = DraftTitleTargetLen - TitleMarginLen
	DraftTitleMaxLen    = DraftTitleTargetLen + TitleMarginLen

	// topic.title（見出し。朗読されず全体文字数にも数えない）。target ± margin の畳み込み。
	DraftTopicTitleTarget = 20
	TopicTitleMarginLen   = 10
	DraftTopicTitleMinLen = DraftTopicTitleTarget - TopicTitleMarginLen
	DraftTopicTitleMaxLen = DraftTopicTitleTarget + TopicTitleMarginLen

	// intro（秒 × CharsPerSecond）。
	DraftIntroMinLen = DraftIntroMinSec * CharsPerSecond
	DraftIntroTarget = DraftIntroTgtSec * CharsPerSecond
	DraftIntroMaxLen = DraftIntroMaxSec * CharsPerSecond

	// closingSummary（秒 × CharsPerSecond）。
	DraftClosingMinLen = DraftClosingMinSec * CharsPerSecond
	DraftClosingTarget = DraftClosingTgtSec * CharsPerSecond
	DraftClosingMaxLen = DraftClosingMaxSec * CharsPerSecond

	// topic.preface（秒 × CharsPerSecond）。
	DraftTopicPrefaceMinLen = DraftTopicPrefaceMinSec * CharsPerSecond
	DraftTopicPrefaceTarget = DraftTopicPrefaceTgtSec * CharsPerSecond
	DraftTopicPrefaceMaxLen = DraftTopicPrefaceMaxSec * CharsPerSecond

	// topic.detail（秒 × CharsPerSecond）。
	DraftTopicDetailMinLen = DraftTopicDetailMinSec * CharsPerSecond
	DraftTopicDetailTarget = DraftTopicDetailTgtSec * CharsPerSecond
	DraftTopicDetailMaxLen = DraftTopicDetailMaxSec * CharsPerSecond

	// topic 数（秒非依存）。固定値のみ。Min/Max は持たない
	// （topic 数を変えても他の定数は壊れない設計にしたため、range 検証は不要）。
	DraftTopicCountTarget = 5

	// SourceItemsPerTopic は情報源 1 件あたりの ItemSource 取得件数上限の目安（各 Adapter の
	// 既定上限 20 件 ÷ DraftTopicCountTarget 5）。system-test が topicCount を絞る時、
	// source 取得件数もこの比率で追従させる（本番は各 Adapter の既定上限を使うため参照しない）。
	SourceItemsPerTopic = 4

	// 全体文字数（挨拶を除く朗読 field の合計。秒 × CharsPerSecond）。
	DraftTotalCharsMin    = DraftTotalMinSec * CharsPerSecond
	DraftTotalCharsTarget = DraftTotalTgtSec * CharsPerSecond
	DraftTotalCharsMax    = DraftTotalMaxSec * CharsPerSecond

	DraftSentenceSuffixRune = '。'
)

// TotalCharsMinFor / TotalCharsTargetFor / TotalCharsMaxFor は topicCount 件の topic を
// 前提にした全体文字数の Min/Target/Max を返す。DraftTotalCharsMin/Target/Max の一般化。
// manuscript_draft_seconds.go の TotalMinSecFor 等 × CharsPerSecond の畳み込み。
//
// @require topicCount >= 0。
// @ensure TotalCharsMinFor(DraftTopicCountTarget) == DraftTotalCharsMin（Target/Max も同様）。
func TotalCharsMinFor(topicCount int) int {
	return TotalMinSecFor(topicCount) * CharsPerSecond
}

func TotalCharsTargetFor(topicCount int) int {
	return TotalTgtSecFor(topicCount) * CharsPerSecond
}

func TotalCharsMaxFor(topicCount int) int {
	return TotalMaxSecFor(topicCount) * CharsPerSecond
}
