package build_test

// このファイルは wav_duration / wav_concat の両 sociable unit test が共有する
// fixture 合成 helper 群を集約する中立な置き場（case は持たない）。
// testdata/ を使わず test 内で PCM WAV bytes を合成する方針に沿い、
// helper はすべて RIFF 仕様の bytes を直接組み、production の buildWAV には依存させない。
// assertion helper（assertCorruptSpeechAudio）と data chunk 走査 helper も
// 両 file が使うためここへ置く。

import (
	"encoding/binary"
	stderrors "errors"
	"math"
	"testing"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

// 合成 WAVE の PCM パラメータ（16-bit PCM / mono / 8000Hz の最小構成）。
const (
	testSampleRate    = 8000
	testChannels      = 1
	testBitsPerSample = 16
)

// synthPCMWav は指定秒数のゼロ埋め data を持つ 44 byte 標準 header の RIFF/WAVE を返す。
func synthPCMWav(t *testing.T, durationSec float64) []byte {
	t.Helper()

	blockAlign := testChannels * testBitsPerSample / 8
	dataLen := int(math.Round(durationSec*float64(testSampleRate))) * blockAlign
	// data 本体はゼロ埋めのまま（無音）。
	return synthPCMWavWithData(t, make([]byte, dataLen))
}

// fmtChunkBytes は 16 byte 標準 fmt chunk（"fmt " id + size + body）を返す。
func fmtChunkBytes() []byte {
	return fmtChunkBytesWithParams(testChannels, testSampleRate, testBitsPerSample, nil)
}

// fmtChunkBytesWithParams は PCM パラメータと余剰 bytes を指定して fmt chunk を組む。
// extra が非 nil なら 16 byte 標準フィールドの後ろへ連結し、拡張 fmt（cbSize 付き等）を表現する。
func fmtChunkBytesWithParams(channels, sampleRate, bitsPerSample int, extra []byte) []byte {
	blockAlign := channels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign

	bodyLen := 16 + len(extra)
	seg := make([]byte, 8+bodyLen)
	copy(seg[0:4], []byte("fmt "))
	binary.LittleEndian.PutUint32(seg[4:8], uint32(bodyLen))
	binary.LittleEndian.PutUint16(seg[8:10], 1) // PCM
	binary.LittleEndian.PutUint16(seg[10:12], uint16(channels))
	binary.LittleEndian.PutUint32(seg[12:16], uint32(sampleRate))
	binary.LittleEndian.PutUint32(seg[16:20], uint32(byteRate))
	binary.LittleEndian.PutUint16(seg[20:22], uint16(blockAlign))
	binary.LittleEndian.PutUint16(seg[22:24], uint16(bitsPerSample))
	copy(seg[24:], extra)
	return seg
}

// dataChunkBytes は "data" id + size + body の bytes を返す。size は body 長そのまま。
func dataChunkBytes(body []byte) []byte {
	seg := make([]byte, 8+len(body))
	copy(seg[0:4], []byte("data"))
	binary.LittleEndian.PutUint32(seg[4:8], uint32(len(body)))
	copy(seg[8:], body)
	return seg
}

// genericChunkBytes は任意 id・任意宣言 size・任意 body の chunk bytes を返す。
// declaredSize と len(body) を独立に指定できるので overrun 系 fixture も組める。
func genericChunkBytes(id string, declaredSize int, body []byte) []byte {
	seg := make([]byte, 8+len(body))
	copy(seg[0:4], []byte(id))
	binary.LittleEndian.PutUint32(seg[4:8], uint32(declaredSize))
	copy(seg[8:], body)
	return seg
}

// wrapRIFF は chunk bytes 群を RIFF/WAVE コンテナで包む。
func wrapRIFF(chunks ...[]byte) []byte {
	return wrapContainer("WAVE", chunks...)
}

// wrapContainer は form type を指定して RIFF コンテナで chunk 群を包む。
func wrapContainer(formType string, chunks ...[]byte) []byte {
	body := []byte{}
	for _, c := range chunks {
		body = append(body, c...)
	}
	buf := make([]byte, 12+len(body))
	copy(buf[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(4+len(body)))
	copy(buf[8:12], []byte(formType))
	copy(buf[12:], body)
	return buf
}

// synthPCMWavWithData は指定 data body を持つ 44 byte 標準 header の RIFF/WAVE を組む。
// fmt(16byte) + data の並びなので data chunk body は固定 offset 44 から始まる。
func synthPCMWavWithData(t *testing.T, data []byte) []byte {
	t.Helper()
	return wrapRIFF(fmtChunkBytes(), dataChunkBytes(data))
}

// assertCorruptSpeechAudio は err が Op = corrupt_speech_audio の Domain Error であることを検証する。
func assertCorruptSpeechAudio(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var de *domainerrors.Error
	if !stderrors.As(err, &de) {
		t.Fatalf("want *errors.Error, got %T (%v)", err, err)
	}
	if de.Op != domainerrors.OpCorruptSpeechAudio {
		t.Fatalf("want Op = %q, got %q", domainerrors.OpCorruptSpeechAudio, de.Op)
	}
}

// readDataChunkOffsetLen は RIFF/WAVE を自前で走査し data chunk body の [開始 offset, 長さ] を返す。
// synthPCMWav / buildWAV の固定 offset 44 に依存せず header を読む。
func readDataChunkOffsetLen(t *testing.T, wav []byte) (int, int) {
	t.Helper()
	if len(wav) < 12 || string(wav[0:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		t.Fatalf("readDataChunk: not a RIFF/WAVE: len=%d", len(wav))
	}
	pos := 12
	for pos+8 <= len(wav) {
		id := string(wav[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(wav[pos+4 : pos+8]))
		body := pos + 8
		if body+size > len(wav) {
			t.Fatalf("readDataChunk: chunk %q overruns buffer", id)
		}
		if id == "data" {
			return body, size
		}
		pos = body + size
		if size%2 == 1 {
			pos++
		}
	}
	t.Fatal("readDataChunk: no data chunk found")
	return 0, 0
}

func readDataChunkLen(t *testing.T, wav []byte) int {
	t.Helper()
	_, n := readDataChunkOffsetLen(t, wav)
	return n
}

func readDataChunkBody(t *testing.T, wav []byte) []byte {
	t.Helper()
	off, n := readDataChunkOffsetLen(t, wav)
	return wav[off : off+n]
}
