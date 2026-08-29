package build

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// ManuscriptDraftFromWriterOutput は TextWriter 戻り string（JSON wire）を ManuscriptDraft へ解釈する。
//
// @require raw は trim 後に非空。wire 形の正本は entities/models.WriterOutput。
// @ensure 成功時は ManuscriptDraft（Title 含む）。失敗時は Domain Error（Op = invalid_manuscript_draft）。
// @ensure Domain Rule の正本は entities/constants/manuscript_draft_limits.go（unmarshal 後に検証）。
// @invariant Infrastructure・vendor envelope を知らない。code fence strip は wire 前処理の最小のみ。
func ManuscriptDraftFromWriterOutput(raw string) (models.ManuscriptDraft, error) {
	body := stripJSONCodeFence(strings.TrimSpace(raw))
	if body == "" {
		return models.ManuscriptDraft{}, draftErr("wire が空")
	}

	var wire models.WriterOutput
	if err := json.Unmarshal([]byte(body), &wire); err != nil {
		return models.ManuscriptDraft{}, errors.DomainErr(errors.OpInvalidManuscriptDraft, err)
	}

	if err := validateWriterOutput(wire); err != nil {
		return models.ManuscriptDraft{}, err
	}

	return toManuscriptDraft(wire), nil
}

// stripJSONCodeFence は先頭末尾の markdown code fence（```json ... ``` / ``` ... ```）を最小 strip する。
// fence が無ければそのまま返す。vendor 固有 envelope は扱わない。
func stripJSONCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	// ```json / ```JSON 等の言語ラベル行（空行含む）を落とす。
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		if isASCIILabelLine(strings.TrimSpace(s[:i])) {
			s = s[i+1:]
		}
	}
	s = strings.TrimSuffix(strings.TrimRight(s, " \t\r\n"), "```")
	return strings.TrimSpace(s)
}

// isASCIILabelLine は s が code fence の言語ラベル行として strip 対象か判定する。
// 空文字（ラベル無し行）または ASCII 英字のみなら true。
func isASCIILabelLine(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}

// validateWriterOutput は manuscript_draft_limits の定数どおりに wire を検証する。
// 失敗はすべて Domain Error（Op = invalid_manuscript_draft）。
func validateWriterOutput(w models.WriterOutput) error {
	if err := validateNarrationField("title", w.Title, constants.DraftTitleMinLen, constants.DraftTitleMaxLen); err != nil {
		return err
	}
	if err := validateNarrationField("intro", w.Intro, constants.DraftIntroMinLen, constants.DraftIntroMaxLen); err != nil {
		return err
	}

	if n := len(w.Topics); n < constants.DraftTopicCountMin || n > constants.DraftTopicCountMax {
		return draftErr(fmt.Sprintf("topic 数が範囲外: %d（%d〜%d）", n, constants.DraftTopicCountMin, constants.DraftTopicCountMax))
	}

	for i, tp := range w.Topics {
		if err := validateNarrationField(fmt.Sprintf("topic[%d].title", i), tp.Title, constants.DraftTopicTitleMinLen, constants.DraftTopicTitleMaxLen); err != nil {
			return err
		}
		if err := validateNarrationField(fmt.Sprintf("topic[%d].preface", i), tp.Preface, constants.DraftTopicPrefaceMinLen, constants.DraftTopicPrefaceMaxLen); err != nil {
			return err
		}
		if err := validateNarrationField(fmt.Sprintf("topic[%d].detail", i), tp.Detail, constants.DraftTopicDetailMinLen, constants.DraftTopicDetailMaxLen); err != nil {
			return err
		}
	}

	if err := validateNarrationField("closingSummary", w.ClosingSummary, constants.DraftClosingMinLen, constants.DraftClosingMaxLen); err != nil {
		return err
	}

	return validateTotalChars(w)
}

// validateNarrationField は朗読 field の基本規則 + rune 数 range を検証する。
func validateNarrationField(name, value string, minLen, maxLen int) error {
	if err := checkNarrationBasics(name, value); err != nil {
		return err
	}
	if n := utf8.RuneCountInString(strings.TrimSpace(value)); n < minLen || n > maxLen {
		return draftErr(fmt.Sprintf("%s の文字数が範囲外: %d（%d〜%d）", name, n, minLen, maxLen))
	}
	return nil
}

// checkNarrationBasics は朗読 field の trim 後非空・日本語含有・末尾句点を検証する。
func checkNarrationBasics(name, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return draftErr(fmt.Sprintf("%s が空", name))
	}
	if !containsJapanese(trimmed) {
		return draftErr(fmt.Sprintf("%s に日本語が含まれない", name))
	}
	if last, _ := utf8.DecodeLastRuneInString(trimmed); last != constants.DraftSentenceSuffixRune {
		return draftErr(fmt.Sprintf("%s の末尾が句点でない", name))
	}
	return nil
}

// validateTotalChars は全朗読 field の rune 合計を DraftTotalCharsMin〜Max で検証する。
func validateTotalChars(w models.WriterOutput) error {
	total := utf8.RuneCountInString(strings.TrimSpace(w.Title)) +
		utf8.RuneCountInString(strings.TrimSpace(w.Intro)) +
		utf8.RuneCountInString(strings.TrimSpace(w.ClosingSummary))
	for _, tp := range w.Topics {
		total += utf8.RuneCountInString(strings.TrimSpace(tp.Title))
		total += utf8.RuneCountInString(strings.TrimSpace(tp.Preface))
		total += utf8.RuneCountInString(strings.TrimSpace(tp.Detail))
	}
	if total < constants.DraftTotalCharsMin || total > constants.DraftTotalCharsMax {
		return draftErr(fmt.Sprintf("全体文字数が範囲外: %d（%d〜%d）", total, constants.DraftTotalCharsMin, constants.DraftTotalCharsMax))
	}
	return nil
}

// containsJapanese は s がひらがな・カタカナ・漢字を 1 文字以上含むか判定する。
func containsJapanese(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x3040 && r <= 0x309F: // ひらがな
			return true
		case r >= 0x30A0 && r <= 0x30FF: // カタカナ
			return true
		case r >= 0x3400 && r <= 0x4DBF: // CJK 拡張 A
			return true
		case r >= 0x4E00 && r <= 0x9FFF: // CJK 統合漢字
			return true
		case r >= 0xF900 && r <= 0xFAFF: // CJK 互換漢字
			return true
		}
	}
	return false
}

// toManuscriptDraft は検証済み wire を ManuscriptDraft へ転記する。
func toManuscriptDraft(w models.WriterOutput) models.ManuscriptDraft {
	topics := make([]models.ManuscriptDraftTopic, len(w.Topics))
	for i, tp := range w.Topics {
		topics[i] = models.ManuscriptDraftTopic{
			Title:   tp.Title,
			Preface: tp.Preface,
			Detail:  tp.Detail,
		}
	}
	return models.ManuscriptDraft{
		Title:          w.Title,
		Intro:          w.Intro,
		Topics:         topics,
		ClosingSummary: w.ClosingSummary,
	}
}

// draftErr は invalid_manuscript_draft の Domain Error を日本語 cause 付きで作る。
func draftErr(msg string) error {
	return errors.DomainErr(errors.OpInvalidManuscriptDraft, stderrors.New(msg))
}
