package main

import (
	"context"
	"encoding/binary"
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

func TestEncodeSegmentTextsToMP3_returnsMP3_whenHappyPath(t *testing.T) {
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

	// Then: err なし。SynthesizeAll に texts が渡り、encode 結果を返す
	if err != nil {
		t.Fatalf("EncodeSegmentTextsToMP3: err = %v, want nil", err)
	}
	if string(got) != string(wantMP3) {
		t.Fatalf("mp3 = %q, want %q", got, wantMP3)
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

	// Then: error。mp3 は空
	if err == nil {
		t.Fatal("err = nil, want synthesize error")
	}
	if got != nil {
		t.Fatalf("mp3 = %v, want nil", got)
	}
}

func TestParseSegmentTextsJSON_returnsTexts_whenArray(t *testing.T) {
	// Given: JSON 配列文字列
	raw := `["α","β","γ"]`

	// When: ParseSegmentTextsJSON する
	got, err := ParseSegmentTextsJSON(raw)

	// Then: 3 要素
	if err != nil {
		t.Fatalf("ParseSegmentTextsJSON: %v", err)
	}
	if len(got) != 3 || got[1] != "β" {
		t.Fatalf("got = %#v", got)
	}
}
