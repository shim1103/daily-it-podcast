package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

type ProduceEpisode struct {
	fetch        *FetchSourceItems
	textWriter   port.TextWriter
	speech       port.SpeechSynthesizer
	writeEpisode *WriteEpisode
	newEpisodeID func() string
	displayLoc   *time.Location
}

// TextWriterMaxAttempts は ManuscriptDraft 検証失敗時の TextWriter 再試行上限。
// LLM 出力の rune 数・topic 数揺れを吸収する。無限 retry を防ぐ。
const TextWriterMaxAttempts = 5

// NewProduceEpisode は Fetch から WriteEpisode までを束ねる Builder UseCase を返す。
//
// @require fetch != nil かつ textWriter != nil かつ speech != nil かつ writeEpisode != nil かつ newEpisodeID != nil かつ displayLoc != nil
// @ensure 戻りは非 nil。
func NewProduceEpisode(
	fetch *FetchSourceItems,
	textWriter port.TextWriter,
	speech port.SpeechSynthesizer,
	writeEpisode *WriteEpisode,
	newEpisodeID func() string,
	displayLoc *time.Location,
) *ProduceEpisode {
	return &ProduceEpisode{
		fetch:        fetch,
		textWriter:   textWriter,
		speech:       speech,
		writeEpisode: writeEpisode,
		newEpisodeID: newEpisodeID,
		displayLoc:   displayLoc,
	}
}

// Run は Fetch から WriteEpisode までの全日次手順を orchestrate する Builder である。
//
// @require uc != nil かつ uc.fetch != nil かつ uc.textWriter != nil かつ uc.speech != nil かつ uc.writeEpisode != nil かつ uc.newEpisodeID != nil かつ uc.displayLoc != nil。now は CLI 実行時刻（Fetch の since 基準かつ date 暦日化の基準）。
// @ensure Fetch 後 0 件なら Domain Error（Op = no_source_items）。WriteEpisode.Run を呼ばない。
// @ensure build.ComposeBrief(items)（constants Prompt へ SOURCES/数値 placeholder/JSON_EXAMPLE 埋め込み）→ TextWriter.Write + ManuscriptDraftFromWriterOutput を最大 TextWriterMaxAttempts 回（draft 検証成功で打ち切り。Write 自体の error は即 return）→ 注入された表示 Location で now を暦日化 → OpeningGreetingTemplate から Greeting 文案（date 注入）→ ClosingFarewell template から Farewell 文案（date 注入）→ TTS 順（Greeting, Intro, 各 topic の Preface, Detail, ClosingSummary, ClosingFarewell）各 1 Synthesize → build.WavDurationSec / 無音込み累積 startSec・durationSec → build.ConcatWAV → opaque UUID episodeId → 完成 manuscript bytes → WriteEpisode.Run。
// @ensure 途中 error なら WriteEpisode.Run を呼ばない（書込なし）。
// @invariant 所有しない: manuscript.schema.json の Validate（Gate）、vendor / env。Infrastructure 型を知らない。表示タイムゾーンの解決（tzdata I/O）は Composition の責務。監視対象一覧・情報源種類を知らない。string→Draft を Port / Adapter に委譲しない。
func (uc *ProduceEpisode) Run(ctx context.Context, now time.Time) error {
	items, err := uc.fetch.Run(ctx, now)
	if err != nil {
		return err
	}

	brief, err := build.ComposeBrief(items)
	if err != nil {
		return err
	}

	draft, err := uc.writeManuscriptDraft(ctx, brief)
	if err != nil {
		return err
	}

	dateStr, spokenDate := displayDate(now, uc.displayLoc)
	greeting := fmt.Sprintf(constants.OpeningGreetingTemplate, spokenDate)
	farewell := fmt.Sprintf(constants.ClosingFarewell, spokenDate)

	segmentTexts := build.SpeechTexts(greeting, farewell, draft)
	segmentWAVs := make([][]byte, len(segmentTexts))
	segmentDurations := make([]float64, len(segmentTexts))
	for i, text := range segmentTexts {
		audio, err := uc.speech.Synthesize(ctx, text)
		if err != nil {
			return err
		}
		dur, err := build.WavDurationSec(audio.Content)
		if err != nil {
			return err
		}
		segmentWAVs[i] = audio.Content
		segmentDurations[i] = dur
	}

	topicStartSecs, durationSec, err := build.Timeline(segmentDurations, len(draft.Topics))
	if err != nil {
		return err
	}

	concatWAV, err := build.ConcatWAV(segmentWAVs...)
	if err != nil {
		return err
	}

	episodeID := uc.newEpisodeID()
	manuscript, err := build.MarshalManuscript(build.ManuscriptInput{
		EpisodeID:      episodeID,
		Date:           dateStr,
		Title:          draft.Title,
		DurationSec:    durationSec,
		Opening:        greeting,
		Draft:          draft,
		TopicStartSecs: topicStartSecs,
		Closing:        draft.ClosingSummary,
	})
	if err != nil {
		return err
	}

	return uc.writeEpisode.Run(ctx, episodeID, manuscript, models.SpeechAudio{Content: concatWAV})
}

// writeManuscriptDraft は TextWriter を最大 TextWriterMaxAttempts 回呼び、valid な ManuscriptDraft を得る。
// 2 回目以降は前回の draft 検証 error を brief 末尾へ付け、同じ失敗の再発を減らす。
func (uc *ProduceEpisode) writeManuscriptDraft(ctx context.Context, brief string) (models.ManuscriptDraft, error) {
	attemptBrief := brief
	var lastErr error
	for attempt := 1; attempt <= TextWriterMaxAttempts; attempt++ {
		raw, err := uc.textWriter.Write(ctx, attemptBrief)
		if err != nil {
			return models.ManuscriptDraft{}, err
		}
		draft, err := build.ManuscriptDraftFromWriterOutput(raw)
		if err == nil {
			return draft, nil
		}
		lastErr = err
		attemptBrief = brief + "\n\n# Previous attempt rejected\n" + err.Error() +
			"\n上記の検証失敗をすべて解消せよ。topics 件数・各 field 文字数・日本語・末尾句点を満たし、JSON オブジェクトのみを出力せよ。\n"
	}
	return models.ManuscriptDraft{}, lastErr
}

// displayDate は now を表示 Location の暦日へ落とし、原稿 date（YYYY-MM-DD）と読み上げ用日付（YYYY年M月D日）を返す。
func displayDate(now time.Time, loc *time.Location) (dateStr, spokenDate string) {
	local := now.In(loc)
	dateStr = local.Format("2006-01-02")
	spokenDate = fmt.Sprintf("%d年%d月%d日", local.Year(), int(local.Month()), local.Day())
	return dateStr, spokenDate
}
