package gemini

import (
	"encoding/binary"
	"math"
	"testing"
)

// synthHelperWAV は指定パラメータで標準 RIFF/WAVE を組む test 用 helper。
func synthHelperWAV(channels uint16, sampleRate uint32, bitsPerSample uint16, durationSec float64) []byte {
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * uint32(blockAlign)
	dataLen := int(math.Round(durationSec * float64(byteRate)))

	buf := make([]byte, 44+dataLen)
	copy(buf[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataLen))
	copy(buf[8:12], []byte("WAVE"))

	copy(buf[12:16], []byte("fmt "))
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1)
	binary.LittleEndian.PutUint16(buf[22:24], channels)
	binary.LittleEndian.PutUint32(buf[24:28], sampleRate)
	binary.LittleEndian.PutUint32(buf[28:32], byteRate)
	binary.LittleEndian.PutUint16(buf[32:34], blockAlign)
	binary.LittleEndian.PutUint16(buf[34:36], bitsPerSample)

	copy(buf[36:40], []byte("data"))
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataLen))

	return buf
}

func TestWAVDurationSec_returnsDuration_whenValidStandardWAV(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		channels      uint16
		sampleRate    uint32
		bitsPerSample uint16
		durationSec   float64
	}{
		{"24kHz mono 1.0s", 1, 24000, 16, 1.0},
		{"24kHz mono 2.5s", 1, 24000, 16, 2.5},
		{"44.1kHz stereo 0.75s", 2, 44100, 16, 0.75},
		{"16kHz mono 3.0s", 1, 16000, 16, 3.0},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wav := synthHelperWAV(tc.channels, tc.sampleRate, tc.bitsPerSample, tc.durationSec)

			got, err := wavDurationSec(wav)
			if err != nil {
				t.Fatalf("wavDurationSec: %v", err)
			}
			if math.Abs(got-tc.durationSec) > 1e-4 {
				t.Fatalf("duration = %v, want %v", got, tc.durationSec)
			}
		})
	}
}

func TestWAVDurationSec_returnsError_whenBufferShorterThanRIFFHeader(t *testing.T) {
	t.Parallel()

	// Given: 12 byte 未満の byte 列
	short := []byte("RIFF1234WAV")

	// When: 尺を算出する
	_, err := wavDurationSec(short)

	// Then: error
	if err == nil {
		t.Fatal("expected error for short buffer")
	}
}

func TestWAVDurationSec_returnsError_whenMissingRIFFOrWAVEMagic(t *testing.T) {
	t.Parallel()

	// Given: RIFF でないヘッダ
	notRIFF := make([]byte, 44)
	copy(notRIFF[0:4], []byte("NOPE"))
	copy(notRIFF[8:12], []byte("WAVE"))

	if _, err := wavDurationSec(notRIFF); err == nil {
		t.Fatal("expected error when RIFF magic missing")
	}

	// Given: WAVE でないヘッダ
	notWAVE := make([]byte, 44)
	copy(notWAVE[0:4], []byte("RIFF"))
	copy(notWAVE[8:12], []byte("NOPE"))

	if _, err := wavDurationSec(notWAVE); err == nil {
		t.Fatal("expected error when WAVE magic missing")
	}
}

func TestWAVDurationSec_returnsError_whenMissingFmtChunk(t *testing.T) {
	t.Parallel()

	// Given: data chunk だけがあり fmt chunk が欠落している
	buf := make([]byte, 20)
	copy(buf[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(buf[4:8], 12)
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("data"))
	binary.LittleEndian.PutUint32(buf[16:20], 0)

	// When: 尺を算出する
	_, err := wavDurationSec(buf)

	// Then: fmt 欠落 error
	if err == nil {
		t.Fatal("expected error when fmt chunk missing")
	}
}

func TestWAVDurationSec_returnsError_whenMissingDataChunk(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk だけがあり data chunk が欠落している
	buf := make([]byte, 36)
	copy(buf[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(buf[4:8], 28)
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("fmt "))
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1)
	binary.LittleEndian.PutUint16(buf[22:24], 1)
	binary.LittleEndian.PutUint32(buf[24:28], 24000)
	binary.LittleEndian.PutUint32(buf[28:32], 48000)
	binary.LittleEndian.PutUint16(buf[32:34], 2)
	binary.LittleEndian.PutUint16(buf[34:36], 16)

	// When: 尺を算出する
	_, err := wavDurationSec(buf)

	// Then: data 欠落 error
	if err == nil {
		t.Fatal("expected error when data chunk missing")
	}
}

func TestWAVDurationSec_calculatesByteRateFromFmtFields_whenByteRateZeroInHeader(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk 内の byteRate フィールドが 0 だが sampleRate・channels・bitsPerSample が非ゼロ
	wav := synthHelperWAV(1, 24000, 16, 2.0)
	binary.LittleEndian.PutUint32(wav[28:32], 0) // zero byteRate in header

	// When: 尺を算出する
	got, err := wavDurationSec(wav)

	// Then: sampleRate * channels * bitsPerSample / 8 から byteRate を導出
	if err != nil {
		t.Fatalf("wavDurationSec: %v", err)
	}
	if math.Abs(got-2.0) > 1e-4 {
		t.Fatalf("duration = %v, want 2.0", got)
	}
}

func TestWAVDurationSec_returnsError_whenByteRateRemainsZero(t *testing.T) {
	t.Parallel()

	// Given: fmt chunk 内の byteRate も sampleRate も 0
	wav := synthHelperWAV(1, 24000, 16, 1.0)
	binary.LittleEndian.PutUint32(wav[24:28], 0) // zero sampleRate
	binary.LittleEndian.PutUint32(wav[28:32], 0) // zero byteRate

	// When: 尺を算出する
	_, err := wavDurationSec(wav)

	// Then: error
	if err == nil {
		t.Fatal("expected error when byteRate cannot be derived")
	}
}
