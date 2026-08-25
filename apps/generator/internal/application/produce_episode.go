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
// 組版（opening / 各 topic / ending の speakable 組み立て・TTS 順序・WAV 結合・完成 manuscript 生成）は Gate（WriteEpisode）ではなく本 UseCase が所有する。
//
// @require uc != nil かつ uc.fetch != nil かつ uc.textWriter != nil かつ uc.speech != nil かつ uc.writeEpisode != nil。now は CLI 実行時刻（Fetch の since 基準かつ date 暦日化の基準）。
// @ensure 所有する policy: brief 組み立て、TextWriter の string → ManuscriptDraft（manuscriptDraftFromWriterOutput。失敗は Domain Error）、OpeningGreeting+Intro→opening speakable、ClosingSummary+ClosingFarewell→ending speakable、TTS 順序（opening → 各 topic の preface+detail → ending）、wavDurationSec / concatWAV、完成 manuscript bytes 組み立て、WriteEpisode.Run 呼び出し。
// @ensure 手順は FetchSourceItems.Run → brief 組み立て → TextWriter.Write（1 回・string）→ manuscriptDraftFromWriterOutput → 朗読単位ごとに SpeechSynthesizer.Synthesize → 各 WAV 尺算出・累積で startSec / durationSec 確定 → concatWAV → 完成 manuscript bytes → WriteEpisode.Run の順で行う。
// @ensure 途中 error なら WriteEpisode.Run を呼ばない（書込なし）。
// @invariant 所有しない: manuscript.schema.json の Validate（Gate）、vendor / env。Infrastructure 型を知らない。監視対象一覧を知らない。string→Draft を Port / Adapter に委譲しない。
func (uc *ProduceEpisode) Run(ctx context.Context, now time.Time) error {
	panic("produce episode: contract stub; logic is C")
}
