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
	encode       port.WAVToMP3Encoder
	writeEpisode *WriteEpisode
	newEpisodeID func() string
	displayLoc   *time.Location
	progress     port.ProgressReporter
	topicCount   int
}

// NewProduceEpisode は Fetch から WriteEpisode までを束ねる Builder UseCase を返す。
//
// @require fetch != nil かつ lookup != nil かつ textWriter != nil かつ speech != nil かつ encode != nil かつ writeEpisode != nil かつ newEpisodeID != nil かつ displayLoc != nil かつ progress != nil かつ topicCount > 0
// @ensure 戻りは非 nil。
func NewProduceEpisode(
	fetch *FetchSourceItems,
	lookup port.CompletedEpisodeLookup,
	textWriter port.TextWriter,
	speech port.SpeechSynthesizer,
	encode port.WAVToMP3Encoder,
	writeEpisode *WriteEpisode,
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

// Run は Fetch から WriteEpisode までの全日次手順を orchestrate する Builder である。
//
// @require uc != nil かつ uc.fetch != nil かつ uc.lookup != nil かつ uc.textWriter != nil かつ uc.speech != nil かつ uc.encode != nil かつ uc.writeEpisode != nil かつ uc.newEpisodeID != nil かつ uc.displayLoc != nil かつ uc.progress != nil。now は CLI 実行時刻（Fetch の since 基準かつ date 暦日化の基準）。
// @ensure 表示 Location で now を暦日化した date につき CompletedEpisodeLookup.HasPair が true なら、Fetch より前に成功 return（episodeID は空。TextWriter / Speech / WriteEpisode を呼ばない）。
// @ensure HasPair が false なら通常どおり続行する。
// @ensure Fetch 後 0 件なら Domain Error（Op = no_source_items）。episodeID は空。WriteEpisode.Run を呼ばない。
// @ensure build.ComposeBriefWithTemplate(items, constants.TextWriterBriefPrompt, uc.topicCount)（constants Prompt へ SOURCES/数値 placeholder/JSON_EXAMPLE 埋め込み）→ TextWriter.Write(ctx, brief, buildFn) で ManuscriptDraft を得る。buildFn は raw を build.ManuscriptDraftFromWriterOutput(raw, uc.topicCount) へ渡すクロージャ（invalid-draft retry・model 切り替え fallback は TextWriter 実装側の責務。Write の error は即 return）→ OpeningGreetingTemplate から Greeting 文案（date 注入）→ ClosingFarewell template から Farewell 文案（date 注入）→ build.SpeechTexts が返す topic+2 束（texts[0] = greeting+intro、各 topic = preface+detail、末尾 = closingSummary+farewell）を 1 回 SynthesizeAll（WAV列を受け取る。retry 予算は Adapter が束ねる）→ build.WavDurationSec / 無音込み累積 topic startSec・ending startSec・durationSec → build.ConcatWAV → WAVToMP3Encoder.EncodeWAVToMP3（完成 WAV→MP3。Infrastructure Error は素通し）→ opaque UUID episodeId → 完成 manuscript bytes（body.opening.text = texts[0]、body.opening.startSec = 0、body.ending.text = texts[末尾]、body.ending.startSec = ending startSec。topic ごとの preface/detail は分けて書く。束ねるのは TTS へ渡す text だけ）→ WriteEpisode.Run（SpeechAudio.Content は MP3 bytes）。
// @ensure WriteEpisode まで到達したら episodeID を返す（Write 失敗時も発行済み ID を返す）。途中 error（Write 前）なら episodeID は空・WriteEpisode.Run を呼ばない。
// @invariant 所有しない: manuscript.schema.json の Validate（Gate）、vendor / env。Infrastructure 型を知らない。表示タイムゾーンの解決（tzdata I/O）は Composition の責務。監視対象一覧・情報源種類を知らない。string→Draft を Port / Adapter に委譲しない。WriteEpisode 内の同日再チェックは持たない。log 出力（step/detail を progress へ渡すだけ。出力面は持たない）。段階通知は Start（呼び出し直前）と Done（成功後のみ。error 時は Done を呼ばない）の 2 点。二重 log 禁止。
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
		EndingStartSec: endingStartSec,
	})
	if err != nil {
		return "", err
	}

	uc.progress.Start("write_episode")
	err = uc.writeEpisode.Run(ctx, episodeID, manuscript, models.SpeechAudio{Content: mp3})
	if err == nil {
		uc.progress.Done("write_episode", episodeID)
	}
	return episodeID, err
}

// displayDate は now を表示 Location の暦日へ落とし、原稿 date（YYYY-MM-DD）と読み上げ用日付（YYYY年M月D日）を返す。
func displayDate(now time.Time, loc *time.Location) (dateStr, spokenDate string) {
	local := now.In(loc)
	dateStr = local.Format("2006-01-02")
	spokenDate = fmt.Sprintf("%d年%d月%d日", local.Year(), int(local.Month()), local.Day())
	return dateStr, spokenDate
}
