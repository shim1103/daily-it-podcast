package gemini

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// pcmToWAV は Gemini TTS の raw PCM（pcmSampleRate / pcmBitDepth / pcmChannels）を WAV bytes へ変換する。
func pcmToWAV(pcm []byte) ([]byte, error) {
	if len(pcm) == 0 {
		return nil, fmt.Errorf("pcm is empty")
	}
	if len(pcm)%2 != 0 {
		return nil, fmt.Errorf("pcm length %d is not aligned to 16-bit samples", len(pcm))
	}

	var buf bytes.Buffer
	// RIFF header
	buf.WriteString("RIFF")
	chunkSize := uint32(36 + len(pcm))
	_ = binary.Write(&buf, binary.LittleEndian, chunkSize)
	buf.WriteString("WAVE")

	// fmt subchunk
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(pcmChannels))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(pcmSampleRate))
	byteRate := uint32(pcmSampleRate * pcmChannels * pcmBitDepth / 8)
	_ = binary.Write(&buf, binary.LittleEndian, byteRate)
	blockAlign := uint16(pcmChannels * pcmBitDepth / 8)
	_ = binary.Write(&buf, binary.LittleEndian, blockAlign)
	_ = binary.Write(&buf, binary.LittleEndian, uint16(pcmBitDepth))

	// data subchunk
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(pcm)))
	buf.Write(pcm)

	return buf.Bytes(), nil
}
