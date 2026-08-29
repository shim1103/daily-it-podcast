package constants

// ManuscriptDraft / TextWriter 出力の parse・validation 用数値定数（Domain Rule 正本）。
// embed 専用 Prompt 文言は text_writer_brief_prompt.go。数値 placeholder は ComposeBrief が本定数で埋める。
// Decision: docs/decisions/2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md

const (
	DraftTitleMinLen    = 30
	DraftTitleTargetLen = 50
	DraftTitleMaxLen    = 70

	DraftIntroMinLen = 10
	DraftIntroTarget = 20
	DraftIntroMaxLen = 30

	DraftTopicTitleMinLen = 10
	DraftTopicTitleTarget = 20
	DraftTopicTitleMaxLen = 30

	DraftTopicPrefaceMinLen = 10
	DraftTopicPrefaceTarget = 20
	DraftTopicPrefaceMaxLen = 30

	DraftTopicDetailMinLen = 10
	DraftTopicDetailTarget = 20
	DraftTopicDetailMaxLen = 30

	DraftTopicCountMin    = 3
	DraftTopicCountTarget = 5
	DraftTopicCountMax    = 8

	DraftTotalCharsMin    = 800
	DraftTotalCharsTarget = 5000
	DraftTotalCharsMax    = 12000

	DraftSentenceSuffixRune = '。'
)
