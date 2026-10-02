package build

import (
	"encoding/binary"
	"math"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// ConcatWAV は sampleRate・channels・bitsPerSample が同じ WAV を 1 本に結合する。
// 隣接 part の間に entities/constants.SegmentSilenceSec 分の無音を挿入する。
//
// @require parts は 1 つ以上。各要素は非空 WAV。
// @ensure 成功時は非空の結合 WAV。header の値は先頭 part と同じ。無音 insert 分は durationSec / startSec 算定に含める。
// @invariant vendor 定数を使わず header から読む。失敗時は entities/errors.Error（Op = corrupt_speech_audio）を返す。
func ConcatWAV(parts ...[]byte) ([]byte, error) {
	if len(parts) == 0 {
		return nil, corruptSpeechAudio("no parts")
	}

	layouts := make([]wavLayout, len(parts))
	for i, part := range parts {
		if len(part) == 0 {
			return nil, corruptSpeechAudio("part is empty")
		}
		layout, err := parseWAV(part)
		if err != nil {
			return nil, err
		}
		layouts[i] = layout
	}

	head := layouts[0]
	for _, l := range layouts[1:] {
		if l.sampleRate != head.sampleRate ||
			l.channels != head.channels ||
			l.bitsPerSample != head.bitsPerSample {
			return nil, corruptSpeechAudio("sampleRate / channels / bitsPerSample mismatch across parts")
		}
	}

	// parseWAV が blockAlign / sampleRate の非ゼロを保証済み。
	// 隣接 part の間 (N-1 箇所) に挿入する無音 1 区間分のバイト数。sample 境界へ整合させる。
	silenceSamples := int(math.Round(constants.SegmentSilenceSec * float64(head.sampleRate)))
	silenceBytes := silenceSamples * int(head.blockAlign)

	totalData := 0
	for _, l := range layouts {
		totalData += len(l.data)
	}
	totalData += silenceBytes * (len(layouts) - 1)

	data := make([]byte, 0, totalData)
	for i, l := range layouts {
		if i > 0 {
			data = append(data, make([]byte, silenceBytes)...)
		}
		data = append(data, l.data...)
	}

	return buildWAV(head, data), nil
}

// buildWAV は data 本体と fmt の値から 44 byte 標準 header の RIFF/WAVE を組み立てる。
// RIFF size と data size は本体長から再計算する。p は parseWAV を通った layout を前提とする。
func buildWAV(p wavLayout, data []byte) []byte {
	out := make([]byte, 44+len(data))
	copy(out[0:4], "RIFF")
	binary.LittleEndian.PutUint32(out[4:8], uint32(36+len(data)))
	copy(out[8:12], "WAVE")

	copy(out[12:16], "fmt ")
	binary.LittleEndian.PutUint32(out[16:20], 16)
	// warn: ⚠️ format tag は 1 固定。parseWAV は入力の format tag を検査しない。
	binary.LittleEndian.PutUint16(out[20:22], 1)
	binary.LittleEndian.PutUint16(out[22:24], p.channels)
	binary.LittleEndian.PutUint32(out[24:28], p.sampleRate)
	binary.LittleEndian.PutUint32(out[28:32], p.byteRate)
	binary.LittleEndian.PutUint16(out[32:34], p.blockAlign)
	binary.LittleEndian.PutUint16(out[34:36], p.bitsPerSample)

	copy(out[36:40], "data")
	binary.LittleEndian.PutUint32(out[40:44], uint32(len(data)))
	copy(out[44:], data)

	return out
}
