package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// synthesizer は SynthesizeAll だけを持つ狭い依存。
type synthesizer interface {
	SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error)
}

// EncodeSegmentTextsToMP3 は produce の speech〜encode 区間だけを実行して mp3 bytes を返す。
//
// @require texts は topic+2 本（opening / 各 topic / ending）。speech・encode は非 nil。
// @ensure 成功時は非空 mp3。Drive へは出ない。
func EncodeSegmentTextsToMP3(
	ctx context.Context,
	texts []string,
	speech synthesizer,
	encode func(context.Context, []byte) ([]byte, error),
) ([]byte, error) {
	if speech == nil {
		return nil, fmt.Errorf("speech is nil")
	}
	if encode == nil {
		return nil, fmt.Errorf("encode is nil")
	}
	if len(texts) < 3 {
		return nil, fmt.Errorf("segmentTexts は 3 本以上必要（got %d）", len(texts))
	}
	audios, err := speech.SynthesizeAll(ctx, texts)
	if err != nil {
		return nil, err
	}
	segmentWAVs := make([][]byte, len(audios))
	segmentDurations := make([]float64, len(audios))
	for i, audio := range audios {
		dur, err := build.WavDurationSec(audio.Content)
		if err != nil {
			return nil, err
		}
		segmentWAVs[i] = audio.Content
		segmentDurations[i] = dur
	}
	topicCount := len(texts) - 2
	if _, _, _, err := build.Timeline(segmentDurations, topicCount); err != nil {
		return nil, err
	}
	concatWAV, err := build.ConcatWAV(segmentWAVs...)
	if err != nil {
		return nil, err
	}
	mp3, err := encode(ctx, concatWAV)
	if err != nil {
		return nil, err
	}
	if len(mp3) == 0 {
		return nil, fmt.Errorf("encode が空 mp3 を返した")
	}
	return mp3, nil
}

// ParseSegmentTextsJSON は workflow 引数の JSON 配列を []string へ直す。
func ParseSegmentTextsJSON(raw string) ([]string, error) {
	var texts []string
	if err := json.Unmarshal([]byte(raw), &texts); err != nil {
		return nil, fmt.Errorf("segmentTexts JSON: %w", err)
	}
	if len(texts) == 0 {
		return nil, fmt.Errorf("segmentTexts が空")
	}
	return texts, nil
}
