package build_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

func TestConcatWAV_insertsConstantSilence_betweenAdjacentParts(t *testing.T) {
	t.Parallel()

	// Given: 0.5 秒の WAVE 2 本
	a := synthPCMWav(t, 0.5)
	b := synthPCMWav(t, 0.5)

	// When: 結合する
	joined, err := build.ConcatWAV(a, b)

	// Then: data 長は 0.5 秒分 + 0.5 秒分 + SegmentSilenceSec 分
	if err != nil {
		t.Fatalf("ConcatWAV: unexpected error: %v", err)
	}
	want := pcmDataBytes(0.5) + pcmDataBytes(0.5) + pcmDataBytes(constants.SegmentSilenceSec)
	if got := readDataChunkLen(t, joined); got != want {
		t.Fatalf("ConcatWAV: joined data len = %d, want %d", got, want)
	}
}

func TestConcatWAV_joinedDataLenEqualsSumOfPartsAndSilence(t *testing.T) {
	t.Parallel()

	// Given: 0.5 秒と 0.3 秒の WAVE
	a := synthPCMWav(t, 0.5)
	b := synthPCMWav(t, 0.3)

	// When: 結合する
	joined, err := build.ConcatWAV(a, b)
	if err != nil {
		t.Fatalf("ConcatWAV: unexpected error: %v", err)
	}

	// Then: joined の data chunk 実バイト長 = part1 data + part2 data + silenceBytes（test 側で独立計算）
	blockAlign := testChannels * testBitsPerSample / 8
	silenceBytes := int(math.Round(constants.SegmentSilenceSec*float64(testSampleRate))) * blockAlign
	part1Data := int(math.Round(0.5*float64(testSampleRate))) * blockAlign
	part2Data := int(math.Round(0.3*float64(testSampleRate))) * blockAlign
	wantDataLen := part1Data + part2Data + silenceBytes

	gotDataLen := readDataChunkLen(t, joined)
	if gotDataLen != wantDataLen {
		t.Fatalf("ConcatWAV: joined data len = %d, want %d", gotDataLen, wantDataLen)
	}

	// Then: part1 分の直後 silenceBytes 分は全てゼロ（挿入位置のバイト検査）
	data := readDataChunkBody(t, joined)
	for i := part1Data; i < part1Data+silenceBytes; i++ {
		if data[i] != 0 {
			t.Fatalf("ConcatWAV: silence byte at offset %d is %d, want 0", i, data[i])
		}
	}
}

func TestConcatWAV_insertsSilenceOnlyBetweenParts_soCountIsNMinusOne(t *testing.T) {
	t.Parallel()

	// Given: 0.5 秒の WAVE 3 本
	parts := [][]byte{
		synthPCMWav(t, 0.5),
		synthPCMWav(t, 0.5),
		synthPCMWav(t, 0.5),
	}

	// When: 結合する
	joined, err := build.ConcatWAV(parts...)

	// Then: 無音は 2 箇所（part 数 3 - 1）だけ入り、data 長は 3 本分 + 無音 2 区間分
	if err != nil {
		t.Fatalf("ConcatWAV: unexpected error: %v", err)
	}
	want := 3*pcmDataBytes(0.5) + 2*pcmDataBytes(constants.SegmentSilenceSec)
	if got := readDataChunkLen(t, joined); got != want {
		t.Fatalf("ConcatWAV: joined data len = %d, want %d", got, want)
	}
}

func TestConcatWAV_returnsSameDataLen_whenSinglePart(t *testing.T) {
	t.Parallel()

	// Given: 0.7 秒の WAVE 1 本
	only := synthPCMWav(t, 0.7)

	// When: 結合する
	joined, err := build.ConcatWAV(only)

	// Then: 無音は入らず data 長は 0.7 秒分のまま
	if err != nil {
		t.Fatalf("ConcatWAV: unexpected error: %v", err)
	}
	if got, want := readDataChunkLen(t, joined), pcmDataBytes(0.7); got != want {
		t.Fatalf("ConcatWAV: joined data len = %d, want %d", got, want)
	}
}

func TestConcatWAV_returnsCorruptSpeechAudio_whenNoParts(t *testing.T) {
	t.Parallel()

	// Given: parts なし
	// When: 結合する
	_, err := build.ConcatWAV()

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestConcatWAV_returnsCorruptSpeechAudio_whenAPartIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 有効な WAVE と空 slice
	valid := synthPCMWav(t, 0.5)

	// When: 結合する
	_, err := build.ConcatWAV(valid, []byte{})

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestConcatWAV_returnsCorruptSpeechAudio_whenAPartIsNotParseable(t *testing.T) {
	t.Parallel()

	// Given: 有効な WAVE と parse 不能な bytes
	valid := synthPCMWav(t, 0.5)

	// When: 結合する
	_, err := build.ConcatWAV(valid, []byte{1, 2, 3})

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestConcatWAV_returnsCorruptSpeechAudio_whenPCMParamsDiffer(t *testing.T) {
	t.Parallel()

	// Given: 8000Hz の WAVE と、sampleRate を 16000 へ書き換えた WAVE
	a := synthPCMWav(t, 0.5)
	b := synthPCMWav(t, 0.5)
	binary.LittleEndian.PutUint32(b[24:28], 16000)

	// When: 結合する
	_, err := build.ConcatWAV(a, b)

	// Then: best-effort 変換せず Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}
