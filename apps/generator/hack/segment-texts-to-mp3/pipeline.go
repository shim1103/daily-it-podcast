package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// speechTextBundleDelimiter は produce と同じ preface/detail 境界（改行3個）。
const speechTextBundleDelimiter = "\n\n\n"

// synthesizer は SynthesizeAll だけを持つ狭い依存。
type synthesizer interface {
	SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error)
}

// EncodeResult は TTS→timeline→mp3 の成果。
type EncodeResult struct {
	MP3            []byte
	TopicStartSecs []float64
	EndingStartSec float64
	DurationSec    float64
}

// EncodeSegmentTextsToMP3 は produce の speech〜encode 区間だけを実行する。
//
// @require texts は topic+2 本（opening / 各 topic / ending）。speech・encode は非 nil。
// @ensure 成功時は非空 mp3 と Timeline 由来の sec。Drive へは出ない。
func EncodeSegmentTextsToMP3(
	ctx context.Context,
	texts []string,
	speech synthesizer,
	encode func(context.Context, []byte) ([]byte, error),
) (EncodeResult, error) {
	var zero EncodeResult
	if speech == nil {
		return zero, fmt.Errorf("speech is nil")
	}
	if encode == nil {
		return zero, fmt.Errorf("encode is nil")
	}
	if len(texts) < 3 {
		return zero, fmt.Errorf("segmentTexts は 3 本以上必要（got %d）", len(texts))
	}
	audios, err := speech.SynthesizeAll(ctx, texts)
	if err != nil {
		return zero, err
	}
	segmentWAVs := make([][]byte, len(audios))
	segmentDurations := make([]float64, len(audios))
	for i, audio := range audios {
		dur, err := build.WavDurationSec(audio.Content)
		if err != nil {
			return zero, err
		}
		segmentWAVs[i] = audio.Content
		segmentDurations[i] = dur
	}
	topicCount := len(texts) - 2
	topicStarts, endingStart, durationSec, err := build.Timeline(segmentDurations, topicCount)
	if err != nil {
		return zero, err
	}
	concatWAV, err := build.ConcatWAV(segmentWAVs...)
	if err != nil {
		return zero, err
	}
	mp3, err := encode(ctx, concatWAV)
	if err != nil {
		return zero, err
	}
	if len(mp3) == 0 {
		return zero, fmt.Errorf("encode が空 mp3 を返した")
	}
	return EncodeResult{
		MP3:            mp3,
		TopicStartSecs: topicStarts,
		EndingStartSec: endingStart,
		DurationSec:    durationSec,
	}, nil
}

// SegmentTextsFromManuscript は完成稿 json から朗読列を一言一句取り出す。
//
// @require body.opening.text / topics[].preface+detail / body.ending.text がある。
// @ensure 本数は 1 + topics + 1。原稿 text は改変しない。
func SegmentTextsFromManuscript(raw []byte) ([]string, error) {
	var doc struct {
		Body struct {
			Opening struct {
				Text string `json:"text"`
			} `json:"opening"`
			Topics []struct {
				Preface string `json:"preface"`
				Detail  string `json:"detail"`
			} `json:"topics"`
			Ending struct {
				Text string `json:"text"`
			} `json:"ending"`
		} `json:"body"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("manuscript JSON: %w", err)
	}
	if doc.Body.Opening.Text == "" || doc.Body.Ending.Text == "" || len(doc.Body.Topics) < 1 {
		return nil, fmt.Errorf("manuscript: opening/ending/topics が不足")
	}
	texts := make([]string, 0, 2+len(doc.Body.Topics))
	texts = append(texts, doc.Body.Opening.Text)
	for _, tp := range doc.Body.Topics {
		texts = append(texts, tp.Preface+speechTextBundleDelimiter+tp.Detail)
	}
	texts = append(texts, doc.Body.Ending.Text)
	return texts, nil
}

// ApplyManuscriptSecs は durationSec と各 startSec だけを TTS timeline へ書き換える。
//
// @require topicStarts 本数 == topics 本数。raw は完成稿形。
// @ensure 原稿 text / title 等は変えず、sec フィールドだけ更新した JSON bytes を返す。
func ApplyManuscriptSecs(raw []byte, topicStarts []float64, endingStartSec, durationSec float64) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("manuscript JSON: %w", err)
	}
	body, ok := doc["body"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("manuscript: body が無い")
	}
	opening, ok := body["opening"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("manuscript: opening が無い")
	}
	ending, ok := body["ending"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("manuscript: ending が無い")
	}
	topicsAny, ok := body["topics"].([]any)
	if !ok {
		return nil, fmt.Errorf("manuscript: topics が無い")
	}
	if len(topicsAny) != len(topicStarts) {
		return nil, fmt.Errorf("topics 本数 %d != topicStarts %d", len(topicsAny), len(topicStarts))
	}
	doc["durationSec"] = durationSec
	opening["startSec"] = 0.0
	for i, t := range topicsAny {
		tm, ok := t.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("topics[%d] が object でない", i)
		}
		tm["startSec"] = topicStarts[i]
	}
	ending["startSec"] = endingStartSec
	out, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EpisodeIDFromManuscript は完成稿の episodeId を返す。
func EpisodeIDFromManuscript(raw []byte) (string, error) {
	var doc struct {
		EpisodeID string `json:"episodeId"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", fmt.Errorf("manuscript JSON: %w", err)
	}
	if doc.EpisodeID == "" {
		return "", fmt.Errorf("episodeId が空")
	}
	return doc.EpisodeID, nil
}
