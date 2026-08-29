package build

import (
	"strconv"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// embedManuscriptDraftLimits は TextWriterBriefPrompt 内の数値 placeholder を
// manuscript_draft_limits の定数で置換する。
func embedManuscriptDraftLimits(prompt string) string {
	replacer := strings.NewReplacer(
		"{{TITLE_MIN}}", strconv.Itoa(constants.DraftTitleMinLen),
		"{{TITLE_MAX}}", strconv.Itoa(constants.DraftTitleMaxLen),
		"{{TITLE_TARGET}}", strconv.Itoa(constants.DraftTitleTargetLen),

		"{{INTRO_MIN}}", strconv.Itoa(constants.DraftIntroMinLen),
		"{{INTRO_MAX}}", strconv.Itoa(constants.DraftIntroMaxLen),
		"{{INTRO_TARGET}}", strconv.Itoa(constants.DraftIntroTarget),

		"{{TOPIC_TITLE_MIN}}", strconv.Itoa(constants.DraftTopicTitleMinLen),
		"{{TOPIC_TITLE_MAX}}", strconv.Itoa(constants.DraftTopicTitleMaxLen),
		"{{TOPIC_TITLE_TARGET}}", strconv.Itoa(constants.DraftTopicTitleTarget),

		"{{PREFACE_MIN}}", strconv.Itoa(constants.DraftTopicPrefaceMinLen),
		"{{PREFACE_MAX}}", strconv.Itoa(constants.DraftTopicPrefaceMaxLen),
		"{{PREFACE_TARGET}}", strconv.Itoa(constants.DraftTopicPrefaceTarget),

		"{{DETAIL_MIN}}", strconv.Itoa(constants.DraftTopicDetailMinLen),
		"{{DETAIL_MAX}}", strconv.Itoa(constants.DraftTopicDetailMaxLen),
		"{{DETAIL_TARGET}}", strconv.Itoa(constants.DraftTopicDetailTarget),

		"{{TOPIC_COUNT_MIN}}", strconv.Itoa(constants.DraftTopicCountMin),
		"{{TOPIC_COUNT_MAX}}", strconv.Itoa(constants.DraftTopicCountMax),
		"{{TOPIC_COUNT_TARGET}}", strconv.Itoa(constants.DraftTopicCountTarget),

		"{{TOTAL_MIN}}", strconv.Itoa(constants.DraftTotalCharsMin),
		"{{TOTAL_MAX}}", strconv.Itoa(constants.DraftTotalCharsMax),
		"{{TOTAL_TARGET}}", strconv.Itoa(constants.DraftTotalCharsTarget),
	)
	return replacer.Replace(prompt)
}
