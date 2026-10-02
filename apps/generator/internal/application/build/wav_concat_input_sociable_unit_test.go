package build_test

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
)

// ConcatWAV が受け取る WAV の読み取り（parseWAV）を、公開 API の ConcatWAV へ 1 本だけ渡して観測する。

func TestConcatWAV_readsWholeData_whenFmtChunkIsExtended(t *testing.T) {
	t.Parallel()

	// Given: 標準 16 byte の後ろに cbSize(2 byte) を持つ拡張 fmt（size=18）の 1.0 秒 WAVE
	extended := fmtChunkBytesWithParams(testChannels, testSampleRate, testBitsPerSample, []byte{0x00, 0x00})
	wav := wrapRIFF(extended, dataChunkBytes(make([]byte, pcmDataBytes(1.0))))

	// When: 1 本だけ結合する
	joined, err := build.ConcatWAV(wav)

	// Then: 余剰フィールドを無視して data を全て読む
	if err != nil {
		t.Fatalf("ConcatWAV: unexpected error: %v", err)
	}
	if got, want := readDataChunkLen(t, joined), pcmDataBytes(1.0); got != want {
		t.Fatalf("ConcatWAV: joined data len = %d, want %d", got, want)
	}
}

func TestConcatWAV_readsWholeData_whenOddSizedUnknownChunkPrecedesFmt(t *testing.T) {
	t.Parallel()

	// Given: fmt の前に odd size(3) の未知 chunk を挟んだ 1.0 秒 WAVE（pad byte 1 個込み）
	oddChunk := genericChunkBytes("LIST", 3, []byte{0x01, 0x02, 0x03, 0x00})
	wav := wrapRIFF(oddChunk, fmtChunkBytes(), dataChunkBytes(make([]byte, pcmDataBytes(1.0))))

	// When: 1 本だけ結合する
	joined, err := build.ConcatWAV(wav)

	// Then: pad byte を跨いで data に到達する
	if err != nil {
		t.Fatalf("ConcatWAV: unexpected error: %v", err)
	}
	if got, want := readDataChunkLen(t, joined), pcmDataBytes(1.0); got != want {
		t.Fatalf("ConcatWAV: joined data len = %d, want %d", got, want)
	}
}

func TestConcatWAV_returnsCorruptSpeechAudio_whenInputIsMalformed(t *testing.T) {
	t.Parallel()

	data16 := dataChunkBytes(make([]byte, 16))
	// sampleRate は fmt chunk の seg offset 12..16、bitsPerSample は 22..24。
	zeroSampleRate := fmtChunkBytes()
	copy(zeroSampleRate[12:16], []byte{0, 0, 0, 0})
	zeroBits := fmtChunkBytes()
	copy(zeroBits[22:24], []byte{0, 0})

	cases := []struct {
		name string
		wav  []byte
	}{
		{"notRIFF", []byte{1, 2, 3}},
		{"onlyRIFFMagic", []byte("RIFF")},
		{"formTypeIsNotWAVE", wrapContainer("AVI ", fmtChunkBytes(), data16)},
		{"fmtChunkAppearsTwice", wrapRIFF(fmtChunkBytes(), fmtChunkBytes(), data16)},
		{"dataChunkAppearsTwice", wrapRIFF(fmtChunkBytes(), data16, data16)},
		{"dataChunkSizeIsZero", synthPCMWavWithData(t, []byte{})},
		{"fmtChunkIsMissing", wrapRIFF(data16)},
		{"dataChunkIsMissing", wrapRIFF(fmtChunkBytes())},
		{"dataChunkSizeOverrunsBuffer", wrapRIFF(fmtChunkBytes(), genericChunkBytes("data", 4096, make([]byte, 16)))},
		{"fmtSampleRateIsZero", wrapRIFF(zeroSampleRate, data16)},
		{"fmtBitsPerSampleIsZero", wrapRIFF(zeroBits, data16)},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given: tc.wav（RIFF/WAVE として不正な入力）
			// When: 1 本だけ結合する
			_, err := build.ConcatWAV(tc.wav)

			// Then: Op = corrupt_speech_audio の Domain Error
			assertCorruptSpeechAudio(t, err)
		})
	}
}
