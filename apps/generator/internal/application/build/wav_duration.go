package build

import (
	"encoding/binary"
	"errors"
	"fmt"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

// corruptSpeechAudio は Op = corrupt_speech_audio の Domain Error を組む package 内 helper。
func corruptSpeechAudio(msg string) error {
	return domainerrors.DomainErr(domainerrors.OpCorruptSpeechAudio, errors.New(msg))
}

// wavLayout は RIFF/WAVE header から読み取った PCM パラメータと data chunk の位置。
// vendor 定数は使わず、すべて header 由来の値のみ保持する。
type wavLayout struct {
	sampleRate    uint32
	channels      uint16
	bitsPerSample uint16
	byteRate      uint32
	blockAlign    uint16
	data          []byte // data chunk の PCM 本体（header を含まない）
}

// WavDurationSec は RIFF/WAV header と data から再生尺（秒）を返す。
//
// @require wav は非空。
// @ensure 成功時 durationSec >= 0。
// @invariant vendor 定数を使わず header から読む。失敗時は entities/errors.Error（Op = corrupt_speech_audio）を返す。
func WavDurationSec(wav []byte) (float64, error) {
	layout, err := parseWAV(wav)
	if err != nil {
		return 0, err
	}
	// parseWAV が byteRate の非ゼロ、data の非空を保証済み。
	return float64(len(layout.data)) / float64(layout.byteRate), nil
}

// parseWAV は RIFF/WAVE bytes を解析し、fmt / data chunk から wavLayout を組み立てる。
// corrupt input（非空でない / マジック不一致 / chunk 欠落 / chunk 重複 / サイズ不整合 / data 空）は
// Op = corrupt_speech_audio の Domain Error を返す。
func parseWAV(wav []byte) (wavLayout, error) {
	if len(wav) == 0 {
		return wavLayout{}, corruptSpeechAudio("wav is empty")
	}
	if len(wav) < 12 {
		return wavLayout{}, corruptSpeechAudio("wav shorter than RIFF header")
	}
	if string(wav[0:4]) != "RIFF" {
		return wavLayout{}, corruptSpeechAudio("missing RIFF magic")
	}
	if string(wav[8:12]) != "WAVE" {
		return wavLayout{}, corruptSpeechAudio("missing WAVE magic")
	}

	var (
		layout   wavLayout
		haveFmt  bool
		haveData bool
	)

	pos := 12
	for pos+8 <= len(wav) {
		id := string(wav[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(wav[pos+4 : pos+8]))
		body := pos + 8
		if body+size > len(wav) {
			return wavLayout{}, corruptSpeechAudio(fmt.Sprintf("chunk %q size %d overruns buffer", id, size))
		}

		switch id {
		case "fmt ":
			if haveFmt {
				return wavLayout{}, corruptSpeechAudio("duplicate fmt chunk")
			}
			if size < 16 {
				return wavLayout{}, corruptSpeechAudio("fmt chunk shorter than 16 bytes")
			}
			seg := wav[body : body+size]
			layout.channels = binary.LittleEndian.Uint16(seg[2:4])
			layout.sampleRate = binary.LittleEndian.Uint32(seg[4:8])
			layout.byteRate = binary.LittleEndian.Uint32(seg[8:12])
			layout.blockAlign = binary.LittleEndian.Uint16(seg[12:14])
			layout.bitsPerSample = binary.LittleEndian.Uint16(seg[14:16])
			if layout.channels == 0 || layout.sampleRate == 0 || layout.bitsPerSample == 0 {
				return wavLayout{}, corruptSpeechAudio("fmt chunk has zero channels / sampleRate / bitsPerSample")
			}
			if layout.blockAlign == 0 {
				layout.blockAlign = layout.channels * layout.bitsPerSample / 8
			}
			if layout.byteRate == 0 {
				layout.byteRate = layout.sampleRate * uint32(layout.blockAlign)
			}
			if layout.blockAlign == 0 || layout.byteRate == 0 {
				return wavLayout{}, corruptSpeechAudio("fmt chunk yields zero blockAlign / byteRate")
			}
			haveFmt = true
		case "data":
			if haveData {
				return wavLayout{}, corruptSpeechAudio("duplicate data chunk")
			}
			if size == 0 {
				return wavLayout{}, corruptSpeechAudio("empty data chunk")
			}
			layout.data = wav[body : body+size]
			haveData = true
		}

		pos = body + size
		if size%2 == 1 { // RIFF chunk は 2 byte 境界へ pad される。
			pos++
		}
	}

	if !haveFmt {
		return wavLayout{}, corruptSpeechAudio("missing fmt chunk")
	}
	if !haveData {
		return wavLayout{}, corruptSpeechAudio("missing data chunk")
	}
	return layout, nil
}
