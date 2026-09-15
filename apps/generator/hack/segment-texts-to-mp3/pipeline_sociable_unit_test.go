package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// minimalWAV は 16-bit mono PCM の短い RIFF/WAVE を組む（hack test 用）。
func minimalWAV(t *testing.T, sampleRate uint32, pcmSamples int) []byte {
	t.Helper()
	if pcmSamples < 1 {
		t.Fatalf("pcmSamples = %d, want >= 1", pcmSamples)
	}
	dataLen := pcmSamples * 2
	buf := make([]byte, 44+dataLen)
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataLen))
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1) // PCM
	binary.LittleEndian.PutUint16(buf[22:24], 1) // mono
	binary.LittleEndian.PutUint32(buf[24:28], sampleRate)
	binary.LittleEndian.PutUint32(buf[28:32], sampleRate*2)
	binary.LittleEndian.PutUint16(buf[32:34], 2)
	binary.LittleEndian.PutUint16(buf[34:36], 16)
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataLen))
	return buf
}

type stubSpeech struct {
	texts []string
	out   []models.SpeechAudio
	err   error
}

func (s *stubSpeech) SynthesizeAll(_ context.Context, texts []string) ([]models.SpeechAudio, error) {
	s.texts = append([]string(nil), texts...)
	return s.out, s.err
}

func TestEncodeSegmentTextsToMP3_returnsMP3AndTimeline_whenHappyPath(t *testing.T) {
	// Given: segmentTexts 3 本（opening / topic / ending）。各 WAV は同一 PCM。encode は固定 mp3
	wav := minimalWAV(t, 24000, 100)
	speech := &stubSpeech{out: []models.SpeechAudio{
		{Content: wav},
		{Content: wav},
		{Content: wav},
	}}
	wantMP3 := []byte("mp3-out")
	encode := func(context.Context, []byte) ([]byte, error) { return wantMP3, nil }
	texts := []string{"opening", "topic", "ending"}

	// When: EncodeSegmentTextsToMP3 を呼ぶ
	got, err := EncodeSegmentTextsToMP3(context.Background(), texts, speech, encode)

	// Then: err なし。mp3 と timeline が埋まる。SynthesizeAll に texts が渡る
	if err != nil {
		t.Fatalf("EncodeSegmentTextsToMP3: err = %v, want nil", err)
	}
	if string(got.MP3) != string(wantMP3) {
		t.Fatalf("mp3 = %q, want %q", got.MP3, wantMP3)
	}
	if len(got.TopicStartSecs) != 1 {
		t.Fatalf("TopicStartSecs len = %d, want 1", len(got.TopicStartSecs))
	}
	if got.DurationSec <= 0 {
		t.Fatalf("DurationSec = %v, want > 0", got.DurationSec)
	}
	if len(speech.texts) != 3 || speech.texts[0] != "opening" {
		t.Fatalf("SynthesizeAll texts = %#v, want %#v", speech.texts, texts)
	}
}

func TestEncodeSegmentTextsToMP3_returnsError_whenSynthesizeFails(t *testing.T) {
	// Given: SynthesizeAll が error
	speech := &stubSpeech{err: errors.New("tts down")}
	encode := func(context.Context, []byte) ([]byte, error) {
		t.Fatal("encode は呼ばれない想定")
		return nil, nil
	}

	// When: EncodeSegmentTextsToMP3 を呼ぶ
	got, err := EncodeSegmentTextsToMP3(context.Background(), []string{"a", "b", "c"}, speech, encode)

	// Then: error。結果はゼロ値
	if err == nil {
		t.Fatal("err = nil, want synthesize error")
	}
	if got.MP3 != nil {
		t.Fatalf("mp3 = %v, want nil", got.MP3)
	}
}

func TestApplyManuscriptSecs_updatesOnlySecFields_whenTimelineGiven(t *testing.T) {
	// Given: 完成稿 json。sec は古い値。原稿 text は固定
	raw := []byte(`{
  "episodeId":"ep-1",
  "date":"2026-09-05",
  "title":"題",
  "durationSec":559.28,
  "body":{
    "opening":{"text":"おはよう","startSec":0},
    "topics":[
      {"title":"T1","preface":"前","detail":"本","startSec":38.72},
      {"title":"T2","preface":"前2","detail":"本2","startSec":100}
    ],
    "ending":{"text":"さようなら","startSec":520.64}
  }
}`)
	topicStarts := []float64{10.5, 40.25}
	endingStart := 70.0
	duration := 90.5

	// When: ApplyManuscriptSecs する
	got, err := ApplyManuscriptSecs(raw, topicStarts, endingStart, duration)

	// Then: sec だけ新値。text / title / preface / detail は元のまま
	if err != nil {
		t.Fatalf("ApplyManuscriptSecs: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if doc["durationSec"].(float64) != duration {
		t.Fatalf("durationSec = %v, want %v", doc["durationSec"], duration)
	}
	if doc["title"] != "題" {
		t.Fatalf("title changed: %v", doc["title"])
	}
	body := doc["body"].(map[string]any)
	opening := body["opening"].(map[string]any)
	if opening["text"] != "おはよう" || opening["startSec"].(float64) != 0 {
		t.Fatalf("opening = %#v", opening)
	}
	topics := body["topics"].([]any)
	t0 := topics[0].(map[string]any)
	if t0["title"] != "T1" || t0["preface"] != "前" || t0["detail"] != "本" {
		t.Fatalf("topic0 text changed: %#v", t0)
	}
	if t0["startSec"].(float64) != 10.5 {
		t.Fatalf("topic0 startSec = %v, want 10.5", t0["startSec"])
	}
	t1 := topics[1].(map[string]any)
	if t1["startSec"].(float64) != 40.25 {
		t.Fatalf("topic1 startSec = %v, want 40.25", t1["startSec"])
	}
	ending := body["ending"].(map[string]any)
	if ending["text"] != "さようなら" || ending["startSec"].(float64) != endingStart {
		t.Fatalf("ending = %#v", ending)
	}
}

func TestSegmentTextsFromManuscript_returnsBundles_whenCompletedJSON(t *testing.T) {
	// Given: opening/ending 全文と topic preface/detail
	raw := []byte(`{
  "episodeId":"ep-1",
  "body":{
    "opening":{"text":"挨拶\n\n\n導入","startSec":0},
    "topics":[{"title":"T","preface":"前","detail":"本","startSec":1}],
    "ending":{"text":"まとめ\n\n\n締め","startSec":2}
  }
}`)

	// When: SegmentTextsFromManuscript する
	got, err := SegmentTextsFromManuscript(raw)

	// Then: opening / preface+detail / ending の 3 本
	if err != nil {
		t.Fatalf("SegmentTextsFromManuscript: %v", err)
	}
	want := []string{"挨拶\n\n\n導入", "前\n\n\n本", "まとめ\n\n\n締め"}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("got = %#v, want %#v", got, want)
	}
}
