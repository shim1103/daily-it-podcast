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

// sourceItemsFetcher は取得窓付き Fetch UseCase（application/fetch）の呼び出し面である。
type sourceItemsFetcher interface {
	Run(ctx context.Context, now time.Time) ([]models.SourceItem, error)
}

type ProduceEpisode struct {
	fetch        sourceItemsFetcher
	lookup       port.CompletedEpisodeLookup
	textWriter   port.TextWriter
	speech       port.SpeechSynthesizer
	encode       port.WAVToMP3Encoder
	writeEpisode port.EpisodeWriter
	newEpisodeID func() string
	displayLoc   *time.Location
	progress     port.ProgressReporter
	topicCount   int
}

// NewProduceEpisode は日次 Produce の Builder UseCase を返す。
//
// @require 全依存非 nil。topicCount > 0。
// @ensure 戻りは非 nil。
func NewProduceEpisode(
	fetch sourceItemsFetcher,
	lookup port.CompletedEpisodeLookup,
	textWriter port.TextWriter,
	speech port.SpeechSynthesizer,
	encode port.WAVToMP3Encoder,
	writeEpisode port.EpisodeWriter,
	newEpisodeID func() string,
	displayLoc *time.Location,
	progress port.ProgressReporter,
	topicCount int,
) *ProduceEpisode {
	return &ProduceEpisode{
		fetch:        fetch,
		lookup:       lookup,
		textWriter:   textWriter,
		speech:       speech,
		encode:       encode,
		writeEpisode: writeEpisode,
		newEpisodeID: newEpisodeID,
		displayLoc:   displayLoc,
		progress:     progress,
		topicCount:   topicCount,
	}
}

// Run は日次 episode を構築して永続する。
//
// @require uc の依存は New 契約どおり。now は実行時刻（Fetch 窓と暦日の基準）。
// @ensure 同日ペア済みなら Fetch より前に成功 return（episodeID 空。後続を呼ばない）。
// @ensure Fetch 0 件なら Domain Error（Op=no_source_items）。episodeID 空。Write しない。
// @ensure 成功時は Write まで完了し episodeID を返す。Write 失敗時も発行済み episodeID を返す。Write 前の失敗では episodeID 空。
// @invariant Infrastructure / env / vendor を知らない。schema Validate は writeepisode Gate 側。progress は成功時のみ Done。
func (uc *ProduceEpisode) Run(ctx context.Context, now time.Time) (episodeID string, err error) {
	dateStr, spokenDate := displayDate(now, uc.displayLoc)

	hasPair, err := uc.lookup.HasPair(ctx, dateStr)
	if err != nil {
		return "", err
	}
	if hasPair {
		uc.progress.Done("already_produced", "")
		return "", nil
	}

	uc.progress.Start("fetch_source_items")
	items, err := uc.fetch.Run(ctx, now)
	if err != nil {
		return "", err
	}
	uc.progress.Done("fetch_source_items", fmt.Sprintf("%d件", len(items)))

	uc.progress.Start("compose_brief")
	brief, err := build.ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt, uc.topicCount)
	if err != nil {
		return "", err
	}
	uc.progress.Done("compose_brief", "")

	uc.progress.Start("write_manuscript_draft")
	draft, err := uc.textWriter.Write(ctx, brief, func(raw string) (models.ManuscriptDraft, error) {
		return build.ManuscriptDraftFromWriterOutput(raw, uc.topicCount)
	})
	if err != nil {
		return "", err
	}
	uc.progress.Done("write_manuscript_draft", fmt.Sprintf("%d topics: %s", len(draft.Topics), draft.Title))

	greeting := fmt.Sprintf(constants.OpeningGreetingTemplate, spokenDate)
	farewell := fmt.Sprintf(constants.ClosingFarewell, spokenDate)

	segmentTexts := build.SpeechTexts(greeting, farewell, draft)
	uc.progress.Start("synthesize_speech")
	audios, err := uc.speech.SynthesizeAll(ctx, segmentTexts)
	if err != nil {
		return "", err
	}
	uc.progress.Done("synthesize_speech", fmt.Sprintf("%dセグメント", len(audios)))
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

	uc.progress.Start("build_timeline")
	topicStartSecs, endingStartSec, durationSec, err := build.Timeline(segmentDurations, len(draft.Topics))
	if err != nil {
		return "", err
	}
	uc.progress.Done("build_timeline", fmt.Sprintf("%.0f秒", durationSec))

	uc.progress.Start("concat_wav")
	concatWAV, err := build.ConcatWAV(segmentWAVs...)
	if err != nil {
		return "", err
	}
	uc.progress.Done("concat_wav", fmt.Sprintf("%dバイト", len(concatWAV)))

	uc.progress.Start("encode_wav_to_mp3")
	mp3, err := uc.encode.EncodeWAVToMP3(ctx, concatWAV)
	if err != nil {
		return "", err
	}
	uc.progress.Done("encode_wav_to_mp3", fmt.Sprintf("%dバイト", len(mp3)))

	episodeID = uc.newEpisodeID()
	// why: contracts/manuscript は読み上げ原稿の SSoT。束の先頭・末尾を opening / ending へ入れる。
	manuscript, err := build.MarshalManuscript(build.ManuscriptInput{
		EpisodeID:      episodeID,
		Date:           dateStr,
		Title:          draft.Title,
		DurationSec:    durationSec,
		Opening:        segmentTexts[0],
		Draft:          draft,
		TopicStartSecs: topicStartSecs,
		Ending:         segmentTexts[len(segmentTexts)-1],
		EndingStartSec: endingStartSec,
	})
	if err != nil {
		return "", err
	}

	uc.progress.Start("write_episode")
	err = uc.writeEpisode.Write(ctx, episodeID, manuscript, models.SpeechAudio{Content: mp3})
	if err != nil {
		return episodeID, err
	}
	uc.progress.Done("write_episode", episodeID)
	return episodeID, nil
}

func displayDate(now time.Time, loc *time.Location) (dateStr, spokenDate string) {
	local := now.In(loc)
	dateStr = local.Format("2006-01-02")
	spokenDate = fmt.Sprintf("%d年%d月%d日", local.Year(), int(local.Month()), local.Day())
	return dateStr, spokenDate
}
