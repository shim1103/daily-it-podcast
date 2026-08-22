package gemini

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// fuzzPCMMaxBytes は fuzz input の上限。local resource を無制限に消費しないための境界。
const fuzzPCMMaxBytes = 1 << 20 // 1MiB

// FuzzPCMToWAV は pcmToWAV に対する Go stdlib fuzz target。
// 任意 byte input で panic しないこと、empty/odd-length は error、
// aligned な非空 input は WAV header と data subchunk の整合を検証する。
func FuzzPCMToWAV(f *testing.F) {
	// Given: 空、奇数長、16-bit aligned な PCM を seed corpus とする
	f.Add([]byte{})
	f.Add([]byte{0x00, 0x01, 0x02})
	f.Add(minimalPCM())

	f.Fuzz(func(t *testing.T, pcm []byte) {
		// Given: input size が上限を超える場合は resource 消費を避けて早期終了する
		if len(pcm) > fuzzPCMMaxBytes {
			t.Skip("pcm exceeds fuzz size limit")
		}

		// When: WAV へ変換する（panic すれば fuzzing が自動で failure とする）
		got, err := pcmToWAV(pcm)

		// Then: empty または奇数長は error
		if len(pcm) == 0 || len(pcm)%2 != 0 {
			if err == nil {
				t.Fatalf("expected error for len(pcm)=%d, got nil", len(pcm))
			}
			return
		}

		// Then: aligned な非空 input は error なしで WAV header と payload が input と整合する
		if err != nil {
			t.Fatalf("pcmToWAV: unexpected error for aligned pcm len=%d: %v", len(pcm), err)
		}
		if !isWAV(got) {
			t.Fatalf("output head = % x, want RIFF/WAVE", got[:min(12, len(got))])
		}

		const headerSize = 44 // RIFF(12) + fmt subchunk(24) + data subchunk header(8)
		if len(got) != headerSize+len(pcm) {
			t.Fatalf("output length = %d, want %d", len(got), headerSize+len(pcm))
		}

		dataLen := binary.LittleEndian.Uint32(got[40:44])
		if int(dataLen) != len(pcm) {
			t.Fatalf("data subchunk length = %d, want %d", dataLen, len(pcm))
		}

		if !bytes.Equal(got[headerSize:], pcm) {
			t.Fatal("payload does not match input pcm")
		}
	})
}
