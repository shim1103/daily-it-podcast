package application

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

type ProduceEpisode struct {
	fetch        *FetchSourceItems
	textWriter   port.TextWriter
	speech       port.SpeechSynthesizer
	writeEpisode *WriteEpisode
}

// NewProduceEpisode は Fetch から WriteEpisode までを束ねる Builder UseCase を返す。
//
// @require fetch != nil かつ textWriter != nil かつ speech != nil かつ writeEpisode != nil
// @ensure 戻りは非 nil。
func NewProduceEpisode(
	fetch *FetchSourceItems,
	textWriter port.TextWriter,
	speech port.SpeechSynthesizer,
	writeEpisode *WriteEpisode,
) *ProduceEpisode {
	return &ProduceEpisode{
		fetch:        fetch,
		textWriter:   textWriter,
		speech:       speech,
		writeEpisode: writeEpisode,
	}
}

// Run は Fetch から WriteEpisode までの全日次手順を orchestrate する Builder である。
//
// @require uc != nil かつ uc.fetch != nil かつ uc.textWriter != nil かつ uc.speech != nil かつ uc.writeEpisode != nil。now は CLI 実行時刻（Fetch の since 基準かつ date 暦日化の基準）。
// @ensure Fetch 後 0 件なら Domain Error（Op = no_source_items）。WriteEpisode.Run を呼ばない。
// @ensure build.ComposeBrief(items)（constants Prompt へ SOURCES/数値 placeholder/JSON_EXAMPLE 埋め込み）→ TextWriter.Write（1 回）→ build.ManuscriptDraftFromWriterOutput（JSON wire）→ JST date 確定 → OpeningGreetingTemplate から Greeting 文案 → TTS 順（Greeting, Intro, 各 topic の Preface, Detail, ClosingSummary, ClosingFarewell）各 1 Synthesize → build.WavDurationSec / 無音込み累積 startSec・durationSec → build.ConcatWAV → opaque UUID episodeId → 完成 manuscript bytes → WriteEpisode.Run。
// @ensure 途中 error なら WriteEpisode.Run を呼ばない（書込なし）。
// @invariant 所有しない: manuscript.schema.json の Validate（Gate）、vendor / env。Infrastructure 型を知らない。監視対象一覧・情報源種類を知らない。string→Draft を Port / Adapter に委譲しない。
func (uc *ProduceEpisode) Run(ctx context.Context, now time.Time) error {
	panic("produce episode: contract stub; logic is D")
}
