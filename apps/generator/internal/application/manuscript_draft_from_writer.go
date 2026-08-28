package application

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

// manuscriptDraftFromWriterOutput は TextWriter の text 断片を ManuscriptDraft へ解釈する。
//
// @require raw は trim 後に非空であってよいが、解釈規則は C で確定する。
// @ensure 成功時は Builder が使える ManuscriptDraft。失敗時は Domain Error（entities/errors.Error, Op = invalid_manuscript_draft）。
// @invariant Infrastructure・vendor envelope を知らない。
func manuscriptDraftFromWriterOutput(raw string) (models.ManuscriptDraft, error) {
	panic("manuscript draft from writer output: contract stub; logic is C")
}
