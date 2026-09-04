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
	lookup       port.CompletedEpisodeLookup
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
// @require fetch != nil かつ lookup != nil かつ textWriter != nil かつ speech != nil かつ writeEpisode != nil かつ newEpisodeID != nil かつ displayLoc != nil
// @ensure 戻りは非 nil。
func NewProduceEpisode(
	fetch *FetchSourceItems,
	lookup port.CompletedEpisodeLookup,
	textWriter port.TextWriter,
	speech port.SpeechSynthesizer,
	writeEpisode *WriteEpisode,
	newEpisodeID func() string,
	displayLoc *time.Location,
) *ProduceEpisode {
	return &ProduceEpisode{
		fetch:        fetch,
		lookup:       lookup,
		textWriter:   textWriter,
		speech:       speech,
		writeEpisode: writeEpisode,
		newEpisodeID: newEpisodeID,
		displayLoc:   displayLoc,
	}
}

// Run は Fetch から WriteEpisode までの全日次手順を orchestrate する Builder である。
//
// @require uc != nil かつ uc.fetch != nil かつ uc.lookup != nil かつ uc.textWriter != nil かつ uc.speech != nil かつ uc.writeEpisode != nil かつ uc.newEpisodeID != nil かつ uc.displayLoc != nil。now は CLI 実行時刻（Fetch の since 基準かつ date 暦日化の基準）。
// @ensure 表示 Location で now を暦日化した date につき CompletedEpisodeLookup.HasPair が true なら、Fetch より前に成功 return（episodeID は空。TextWriter / Speech / WriteEpisode を呼ばない）。
// @ensure HasPair が false なら通常どおり続行する。
// @ensure Fetch 後 0 件なら Domain Error（Op = no_source_items）。episodeID は空。WriteEpisode.Run を呼ばない。
// @ensure build.ComposeBrief(items)（constants Prompt へ SOURCES/数値 placeholder/JSON_EXAMPLE 埋め込み）→ TextWriter.Write + ManuscriptDraftFromWriterOutput を最大 TextWriterMaxAttempts 回（draft 検証成功で打ち切り。Write 自体の error は即 return）→ OpeningGreetingTemplate から Greeting 文案（date 注入）→ ClosingFarewell template から Farewell 文案（date 注入）→ build.SpeechTexts が返す topic+2 束（texts[0] = greeting+intro、各 topic = preface+detail、末尾 = closingSummary+farewell）を 1 回 SynthesizeAll（WAV列を受け取る。retry 予算は Adapter が束ねる）→ build.WavDurationSec / 無音込み累積 startSec・durationSec → build.ConcatWAV → opaque UUID episodeId → 完成 manuscript bytes（body.opening = texts[0]、body.ending = texts[末尾]。topic ごとの preface/detail は分けて書く）→ WriteEpisode.Run。
// @ensure WriteEpisode まで到達したら episodeID を返す（Write 失敗時も発行済み ID を返す）。途中 error（Write 前）なら episodeID は空・WriteEpisode.Run を呼ばない。
// @invariant 所有しない: manuscript.schema.json の Validate（Gate）、vendor / env。Infrastructure 型を知らない。表示タイムゾーンの解決（tzdata I/O）は Composition の責務。監視対象一覧・情報源種類を知らない。string→Draft を Port / Adapter に委譲しない。WriteEpisode 内の同日再チェックは持たない。
func (uc *ProduceEpisode) Run(ctx context.Context, now time.Time) (episodeID string, err error) {
	dateStr, spokenDate := displayDate(now, uc.displayLoc)

	hasPair, err := uc.lookup.HasPair(ctx, dateStr)
	if err != nil {
		return "", err
	}
	if hasPair {
		return "", nil
	}

	items, err := uc.fetch.Run(ctx, now)
	if err != nil {
		return "", err
	}

	brief, err := build.ComposeBrief(items)
	if err != nil {
		return "", err
	}

	draft, err := uc.writeManuscriptDraft(ctx, brief)
	if err != nil {
		return "", err
	}

	greeting := fmt.Sprintf(constants.OpeningGreetingTemplate, spokenDate)
	farewell := fmt.Sprintf(constants.ClosingFarewell, spokenDate)

	segmentTexts := build.SpeechTexts(greeting, farewell, draft)
	audios, err := uc.speech.SynthesizeAll(ctx, segmentTexts)
	if err != nil {
		return "", err
	}
	segmentWAVs := make([][]byte, len(audios))
	segmentDurations := make([]float64, len(audios))
	for i, audio := range audios {
		dur, err := build.WavDurationSec(audio.Content)
		if err != nil {
			return "", err
		}
		segmentWAVs[i] = audio.Content
		segmentDurations[i] = dur
	}

	topicStartSecs, durationSec, err := build.Timeline(segmentDurations, len(draft.Topics))
	if err != nil {
		return "", err
	}

	concatWAV, err := build.ConcatWAV(segmentWAVs...)
	if err != nil {
		return "", err
	}

	episodeID = uc.newEpisodeID()
	// why: contracts/manuscript は TTS が読む原稿そのものの SSoT。読み上げ束の先頭・末尾を body.opening / body.ending へそのまま入れる。
	manuscript, err := build.MarshalManuscript(build.ManuscriptInput{
		EpisodeID:      episodeID,
		Date:           dateStr,
		Title:          draft.Title,
		DurationSec:    durationSec,
		Opening:        segmentTexts[0],
		Draft:          draft,
		TopicStartSecs: topicStartSecs,
		Ending:         segmentTexts[len(segmentTexts)-1],
	})
	if err != nil {
		return "", err
	}

	err = uc.writeEpisode.Run(ctx, episodeID, manuscript, models.SpeechAudio{Content: concatWAV})
	return episodeID, err
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
