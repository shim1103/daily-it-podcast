package build

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

// ManuscriptDraftFromWriterOutput は TextWriter 戻り string（JSON wire）を ManuscriptDraft へ解釈する。
//
// @require raw は trim 後に非空。wire 形の正本は entities/models.WriterOutput。
// @ensure 成功時は ManuscriptDraft（Title 含む）。失敗時は Domain Error（Op = invalid_manuscript_draft）。
// @ensure Domain Rule の正本は entities/constants/manuscript_draft_limits.go（unmarshal 後に検証）。
// @invariant Infrastructure・vendor envelope を知らない。
func ManuscriptDraftFromWriterOutput(raw string) (models.ManuscriptDraft, error) {
	panic("manuscript draft from writer output: contract stub; logic is C")
}
