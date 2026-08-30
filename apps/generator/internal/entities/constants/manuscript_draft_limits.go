package constants

// ManuscriptDraft / TextWriter 出力の parse・validation 用文字数定数（Domain Rule 正本）。
// 朗読 field（intro / closingSummary / topic.preface / topic.detail / 全体）の文字数は
// manuscript_draft_seconds.go の秒数 × CharsPerSecond の const 畳み込みで導出する。
// title / topic.title は朗読されない見出しなので秒非依存の文字数を直接定義する。
// embed 専用 Prompt 文言は text_writer_brief_prompt.go。数値 placeholder は ComposeBrief が本定数で埋める。
// Decision: docs/decisions/2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md

const (
	// title（見出し。朗読されず全体文字数にも数えない）。
	DraftTitleMinLen    = 30
	DraftTitleTargetLen = 50
	DraftTitleMaxLen    = 70

	// topic.title（見出し。朗読されず全体文字数にも数えない）。
	DraftTopicTitleMinLen = 10
	DraftTopicTitleTarget = 20
	DraftTopicTitleMaxLen = 30

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

	// topic 数（秒非依存）。
	DraftTopicCountMin    = 3
	DraftTopicCountTarget = 5
	DraftTopicCountMax    = 7

	// 全体文字数（挨拶を除く朗読 field の合計。秒 × CharsPerSecond）。
	DraftTotalCharsMin    = DraftTotalMinSec * CharsPerSecond
	DraftTotalCharsTarget = DraftTotalTgtSec * CharsPerSecond
	DraftTotalCharsMax    = DraftTotalMaxSec * CharsPerSecond

	DraftSentenceSuffixRune = '。'
)
