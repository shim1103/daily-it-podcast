// Package manuscript は原稿取得の UseCase を提供する。
// source を順に試し、番兵 error の種類に応じて次 source へ渡す brief を変えて切り替える。
package manuscript

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.TextWriter = (*TextWriter)(nil)

// why: event 名の typo は compile で捕まらないので定数化する。
const fallbackEventSourceSwitched = "manuscript_source_switched"

// TextWriter は sources を先頭から順に試し原稿断片を確保する UseCase である。
type TextWriter struct {
	sources  []port.TextWriter
	fallback port.FallbackReporter
}

// NewTextWriter は sources と、切り替え発生時の観測面を束ねる。
//
// @require len(sources) > 0 かつ全要素が非 nil。fallback != nil。
// @ensure 戻りは *TextWriter（port.TextWriter を満たす）。
func NewTextWriter(sources []port.TextWriter, fallback port.FallbackReporter) *TextWriter {
	return &TextWriter{sources: sources, fallback: fallback}
}

// Write は sources を先頭から順に試す。
//
// @require brief は trim 後に非空。違反時は domainerrors.DomainErr(OpEmptyBrief) を返し source を呼ばない。
// @require buildFn は非 nil。各 source.Write へそのまま渡す。
// @ensure ある source が成功したら即その draft を返し、以降の source を呼ばない。
// @ensure source の error が errors.Is(err, port.ErrSourceExhausted)==true のとき、fallback.Fallback を呼んでから
//
//	次 source へ切り替える。error chain から port.LastAttempt を取り出せれば
//	（errors.As）port.BuildRejectionBrief の brief で、取り出せなければ「素の brief」で切り替える。
//
// @ensure source の error が errors.Is(err, port.ErrDraftRejected)==true のとき、fallback.Fallback を呼んでから
//
//	次 source へ切り替える。port.LastAttempt を取り出せれば port.BuildRejectionBrief の brief、
//	取り出せなければ port.BuildRejectionBriefWithoutRaw の brief で切り替える。
//
// @ensure source の error がどちらの番兵でもないとき、その error をそのまま返し以降の source を呼ばない。
// @ensure 全 source を使い切ったら最後の source の error を返す。
// @invariant sources の順序は呼び出し順を規定する（追い越しなし）。切り替え時に次 source の error を
//
//	別型で再 wrap しない。番兵の判定は種別のみで vendor を問わない。invalid-draft retry 自体は
//	持たない（各 source の Adapter 実装が持つ）。
func (w *TextWriter) Write(ctx context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		// why: 前提違反は Application 層の Domain Error 規約（write_episode の OpEmptyEpisodeID 等）に揃える。
		return models.ManuscriptDraft{}, domainerrors.DomainErr(domainerrors.OpEmptyBrief, fmt.Errorf("brief is empty after trim"))
	}

	attemptBrief := brief
	var lastErr error
	for _, source := range w.sources {
		draft, err := source.Write(ctx, attemptBrief, buildFn)
		if err == nil {
			return draft, nil
		}
		lastErr = err

		isDraftRejected := errors.Is(err, port.ErrDraftRejected)
		if !errors.Is(err, port.ErrSourceExhausted) && !isDraftRejected {
			return models.ManuscriptDraft{}, err
		}
		w.fallback.Fallback(fallbackEventSourceSwitched)

		var lastAttempt port.LastAttempt
		switch {
		case errors.As(err, &lastAttempt):
			attemptBrief = port.BuildRejectionBrief(brief, lastAttempt.Raw, lastAttempt.BuildErr.Error())
		case isDraftRejected:
			attemptBrief = port.BuildRejectionBriefWithoutRaw(brief, err.Error())
		default:
			attemptBrief = brief
		}
	}

	return models.ManuscriptDraft{}, lastErr
}
