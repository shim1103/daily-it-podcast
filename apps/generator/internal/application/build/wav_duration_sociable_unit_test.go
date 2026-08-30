package build_test

import (
	"math"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
)

func TestWavDurationSec_returnsDuration_whenSynthesizedWAVHasKnownLength(t *testing.T) {
	t.Parallel()

	// Given: 既知尺の合成 WAVE 群（0.25s / 1.0s / 2.5s）
	cases := []float64{0.25, 1.0, 2.5}

	for _, want := range cases {
		want := want
		wav := synthPCMWav(t, want)

		// When: 再生尺を求める
		got, err := build.WavDurationSec(wav)

		// Then: 誤差許容で一致
		if err != nil {
			t.Fatalf("WavDurationSec(%.2fs): unexpected error: %v", want, err)
		}
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("WavDurationSec(%.2fs): want ~%v, got %v", want, want, got)
		}
	}
}

func TestWavDurationSec_returnsDuration_whenFmtChunkIsExtended(t *testing.T) {
	t.Parallel()

	// Given: 標準 16 byte の後ろに cbSize(2byte) を持つ拡張 fmt（size=18）の 1.0 秒 WAVE
	extended := fmtChunkBytesWithParams(testChannels, testSampleRate, testBitsPerSample, []byte{0x00, 0x00})
	blockAlign := testChannels * testBitsPerSample / 8
	dataLen := int(math.Round(1.0*float64(testSampleRate))) * blockAlign
	wav := wrapRIFF(extended, dataChunkBytes(make([]byte, dataLen)))

	// When: 再生尺を求める
	got, err := build.WavDurationSec(wav)

	// Then: 余剰フィールドを無視して 1.0 秒
	if err != nil {
		t.Fatalf("WavDurationSec: unexpected error: %v", err)
	}
	if math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("WavDurationSec: want ~1.0, got %v", got)
	}
}

func TestWavDurationSec_returnsDuration_whenOddSizedUnknownChunkPrecedesFmt(t *testing.T) {
	t.Parallel()

	// Given: fmt の前に odd size(3) の未知 chunk を挟んだ 1.0 秒 WAVE（pad byte 1 個込み）
	oddChunk := genericChunkBytes("LIST", 3, []byte{0x01, 0x02, 0x03, 0x00}) // body 4 byte（3 + pad 1）
	blockAlign := testChannels * testBitsPerSample / 8
	dataLen := int(math.Round(1.0*float64(testSampleRate))) * blockAlign
	wav := wrapRIFF(oddChunk, fmtChunkBytes(), dataChunkBytes(make([]byte, dataLen)))

	// When: 再生尺を求める
	got, err := build.WavDurationSec(wav)

	// Then: pad byte を跨いで data に到達し 1.0 秒
	if err != nil {
		t.Fatalf("WavDurationSec: unexpected error: %v", err)
	}
	if math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("WavDurationSec: want ~1.0, got %v", got)
	}
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenBytesAreNotRIFF(t *testing.T) {
	t.Parallel()

	// Given: RIFF/WAVE として解釈できない bytes
	broken := []byte{1, 2, 3}

	// When: 再生尺を求める
	_, err := build.WavDurationSec(broken)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenOnlyRIFFMagicPresent(t *testing.T) {
	t.Parallel()

	// Given: RIFF マジックだけで chunk を持たない bytes
	riffOnly := []byte("RIFF")

	// When: 再生尺を求める
	_, err := build.WavDurationSec(riffOnly)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenFormTypeIsNotWAVE(t *testing.T) {
	t.Parallel()

	// Given: RIFF だが form type が WAVE でない（AVI ）bytes
	notWave := wrapContainer("AVI ", fmtChunkBytes(), dataChunkBytes(make([]byte, 16)))

	// When: 再生尺を求める
	_, err := build.WavDurationSec(notWave)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenFmtChunkAppearsTwice(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk を 2 つ持つ RIFF/WAVE（RIFF 仕様上 fmt は 1 回のみ）
	wav := wrapRIFF(fmtChunkBytes(), fmtChunkBytes(), dataChunkBytes(make([]byte, 16)))

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: 2 つ目の fmt を検出し Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenDataChunkAppearsTwice(t *testing.T) {
	t.Parallel()

	// Given: data chunk を 2 つ持つ RIFF/WAVE（RIFF 仕様上 data は 1 回のみ）
	wav := wrapRIFF(fmtChunkBytes(), dataChunkBytes(make([]byte, 16)), dataChunkBytes(make([]byte, 16)))

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: 2 つ目の data を検出し Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenDataChunkSizeIsZero(t *testing.T) {
	t.Parallel()

	// Given: data chunk の宣言 size が 0（PCM サンプル無し）の RIFF/WAVE
	wav := synthPCMWavWithData(t, []byte{})

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: duration 0 秒の正常 WAV として通さず Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenFmtChunkIsMissing(t *testing.T) {
	t.Parallel()

	// Given: data chunk のみで fmt chunk を持たない RIFF/WAVE
	wav := wrapRIFF(dataChunkBytes(make([]byte, 16)))

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenDataChunkIsMissing(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk のみで data chunk を持たない RIFF/WAVE
	wav := wrapRIFF(fmtChunkBytes())

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenDataChunkSizeOverrunsBuffer(t *testing.T) {
	t.Parallel()

	// Given: data chunk の宣言 size が実 body 長を超える RIFF/WAVE
	overrun := genericChunkBytes("data", 4096, make([]byte, 16))
	wav := wrapRIFF(fmtChunkBytes(), overrun)

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenFmtSampleRateIsZero(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk の sampleRate フィールドが 0 の RIFF/WAVE
	fmtSeg := fmtChunkBytes()
	// body 先頭は id(4) + size(4) の後ろ。sampleRate は body offset 4..8 → seg offset 12..16。
	for i := 12; i < 16; i++ {
		fmtSeg[i] = 0
	}
	wav := wrapRIFF(fmtSeg, dataChunkBytes(make([]byte, 16)))

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}

func TestWavDurationSec_returnsCorruptSpeechAudio_whenFmtBitsPerSampleIsZero(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk の bitsPerSample フィールドが 0 の RIFF/WAVE
	fmtSeg := fmtChunkBytes()
	// bitsPerSample は body offset 14..16 → seg offset 22..24。
	fmtSeg[22] = 0
	fmtSeg[23] = 0
	wav := wrapRIFF(fmtSeg, dataChunkBytes(make([]byte, 16)))

	// When: 再生尺を求める
	_, err := build.WavDurationSec(wav)

	// Then: Op = corrupt_speech_audio の Domain Error
	assertCorruptSpeechAudio(t, err)
}
