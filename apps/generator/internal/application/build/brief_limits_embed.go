package build

import (
	"strconv"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// embedManuscriptDraftLimits は TextWriterBriefPrompt 内の数値 placeholder を
// manuscript_draft_limits の定数で置換する。topicCount に依存する placeholder
// （{{TOPIC_COUNT_TARGET}} / {{TOTAL_*}} / {{TOTAL_MINUTES_*}}）は
// manuscript_draft_seconds.go / manuscript_draft_limits.go の *For 関数から topicCount で導出する。
//
// @require topicCount >= 0。
// @ensure 戻りは prompt の数値 placeholder をすべて置換した文字列（{{SOURCES}} / {{JSON_EXAMPLE}} は残す）。
func embedManuscriptDraftLimits(prompt string, topicCount int) string {
	totalMinSec, totalMaxSec := constants.TotalMinSecFor(topicCount), constants.TotalMaxSecFor(topicCount)
	replacer := strings.NewReplacer(
		"{{TITLE_MIN}}", strconv.Itoa(constants.DraftTitleMinLen),
		"{{TITLE_MAX}}", strconv.Itoa(constants.DraftTitleMaxLen),
		"{{TITLE_TARGET}}", strconv.Itoa(constants.DraftTitleTargetLen),

		"{{INTRO_MIN}}", strconv.Itoa(constants.DraftIntroMinLen),
		"{{INTRO_MAX}}", strconv.Itoa(constants.DraftIntroMaxLen),
		"{{INTRO_TARGET}}", strconv.Itoa(constants.DraftIntroTarget),

		"{{CLOSING_MIN}}", strconv.Itoa(constants.DraftClosingMinLen),
		"{{CLOSING_MAX}}", strconv.Itoa(constants.DraftClosingMaxLen),
		"{{CLOSING_TARGET}}", strconv.Itoa(constants.DraftClosingTarget),

		"{{TOPIC_TITLE_MIN}}", strconv.Itoa(constants.DraftTopicTitleMinLen),
		"{{TOPIC_TITLE_MAX}}", strconv.Itoa(constants.DraftTopicTitleMaxLen),
		"{{TOPIC_TITLE_TARGET}}", strconv.Itoa(constants.DraftTopicTitleTarget),

		"{{PREFACE_MIN}}", strconv.Itoa(constants.DraftTopicPrefaceMinLen),
		"{{PREFACE_MAX}}", strconv.Itoa(constants.DraftTopicPrefaceMaxLen),
		"{{PREFACE_TARGET}}", strconv.Itoa(constants.DraftTopicPrefaceTarget),

		"{{DETAIL_MIN}}", strconv.Itoa(constants.DraftTopicDetailMinLen),
		"{{DETAIL_MAX}}", strconv.Itoa(constants.DraftTopicDetailMaxLen),
		"{{DETAIL_TARGET}}", strconv.Itoa(constants.DraftTopicDetailTarget),

		"{{TOPIC_COUNT_TARGET}}", strconv.Itoa(topicCount),

		"{{TOTAL_MIN}}", strconv.Itoa(constants.TotalCharsMinFor(topicCount)),
		"{{TOTAL_MAX}}", strconv.Itoa(constants.TotalCharsMaxFor(topicCount)),
		"{{TOTAL_TARGET}}", strconv.Itoa(constants.TotalCharsTargetFor(topicCount)),
		"{{TOTAL_MINUTES_MIN}}", strconv.Itoa(totalMinSec/60),
		"{{TOTAL_MINUTES_MAX}}", strconv.Itoa(totalMaxSec/60),
	)
	return replacer.Replace(prompt)
}
