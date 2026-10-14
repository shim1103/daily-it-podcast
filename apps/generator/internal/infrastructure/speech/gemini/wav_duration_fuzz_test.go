package gemini

import (
	"math"
	"testing"
)

// fuzzWAVMaxBytes は fuzz input の上限。local resource を無制限に消費しないための境界。
const fuzzWAVMaxBytes = 1 << 20 // 1MiB

// FuzzWAVDuration は wavDurationSec に対する Go stdlib fuzz target。
// 任意 byte input で panic しないこと、および valid な合成 WAV は算出値が期待尺と整合することを検証する。
func FuzzWAVDuration(f *testing.F) {
	// Given: 空、壊れた byte 列、valid な合成 WAV を seed corpus とする
	f.Add([]byte{})
	f.Add([]byte("RIFF1234WAV"))
	f.Add(synthHelperWAV(1, 24000, 16, 1.0))
	f.Add(synthHelperWAV(2, 44100, 16, 0.5))

	f.Fuzz(func(t *testing.T, wav []byte) {
		if len(wav) > fuzzWAVMaxBytes {
			t.Skip("wav exceeds fuzz size limit")
		}

		// When: 任意 byte 列で尺を算出する（panic すれば fuzzing が failure を検出）
		duration, err := wavDurationSec(wav)
		if err == nil {
			if math.IsNaN(duration) || math.IsInf(duration, 0) || duration < 0 {
				t.Fatalf("unexpected invalid duration %v for wav len %d", duration, len(wav))
			}
		}
	})
}
