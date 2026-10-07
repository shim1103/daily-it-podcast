// Package manuscript は原稿取得の UseCase を提供する。
// 取得元（source）を順に試し、原稿を得られなかった source があれば次の source へ切り替える。
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
// @ensure brief が trim 後に空なら domainerrors.DomainErr(OpEmptyBrief) を返し、source を呼ばない。
// @ensure ある source が成功したら即その draft を返し、以降の source を呼ばない。
// @ensure source の error が port.ErrSourceExhausted または port.ErrDraftRejected を含むとき、
// fallback.Fallback を呼んでから次 source へ切り替える。次 source へ渡す brief は、error chain から
// port.LastAttempt を取り出せれば port.BuildRejectionBrief、取り出せなければ素の brief。
// @ensure どちらの番兵も含まない error は、そのまま返して以降の source を呼ばない。
// @ensure 全 source が切り替え対象の error なら、最後の source の error を返す。
// @invariant sources の順序は呼び出し順を規定する。切り替え時に error を再 wrap しない。
// invalid-draft retry は持たない（各 source の Adapter が持つ）。
func (w *TextWriter) Write(ctx context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	if strings.TrimSpace(brief) == "" {
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
		if !shouldSwitchSource(err) {
			return models.ManuscriptDraft{}, err
		}
		w.fallback.Fallback(fallbackEventSourceSwitched)
		attemptBrief = nextSourceBrief(brief, err)
		lastErr = err
	}

	return models.ManuscriptDraft{}, lastErr
}

const fallbackEventSourceSwitched = "manuscript_source_switched"

func shouldSwitchSource(err error) bool {
	return errors.Is(err, port.ErrSourceExhausted) || errors.Is(err, port.ErrDraftRejected)
}

func nextSourceBrief(brief string, err error) string {
	var lastAttempt port.LastAttempt
	if errors.As(err, &lastAttempt) {
		return port.BuildRejectionBrief(brief, lastAttempt.Raw, lastAttempt.BuildErr.Error())
	}
	return brief
}
