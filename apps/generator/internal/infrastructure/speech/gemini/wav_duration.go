package gemini

import (
	"encoding/binary"
	"fmt"
)

const (
	riffHeaderLen       = 12
	riffMagicOffset     = 0
	riffFormatOffset    = 8
	riffMagicRIFF       = "RIFF"
	riffFormatWAVE      = "WAVE"
	chunkHeaderLen      = 8
	chunkIDLen          = 4
	minFmtChunkLen      = 16
	fmtChannelsOffset   = 2
	fmtSampleRateOff    = 4
	fmtByteRateOffset   = 8
	fmtBitsPerSampleOff = 14
	bitsPerByte         = 8
	chunkPaddingAlign   = 2
	chunkIDFmt          = "fmt "
	chunkIDData         = "data"
)

// wavDurationSec は RIFF/WAVE header から fmt chunk の byteRate と data chunk のバイト数を読み、再生尺（秒）を算出する。
func wavDurationSec(wav []byte) (float64, error) {
	if err := validateRIFFHeader(wav); err != nil {
		return 0, err
	}
	byteRate, dataSize, err := scanWAVChunks(wav)
	if err != nil {
		return 0, err
	}
	return float64(dataSize) / float64(byteRate), nil
}

func validateRIFFHeader(wav []byte) error {
	if len(wav) < riffHeaderLen {
		return fmt.Errorf("wav is shorter than RIFF header: %d bytes", len(wav))
	}
	if string(wav[riffMagicOffset:riffMagicOffset+chunkIDLen]) != riffMagicRIFF ||
		string(wav[riffFormatOffset:riffFormatOffset+chunkIDLen]) != riffFormatWAVE {
		return fmt.Errorf("missing RIFF/WAVE header")
	}
	return nil
}

func scanWAVChunks(wav []byte) (uint32, uint32, error) {
	var (
		byteRate uint32
		dataSize uint32
		haveFmt  bool
		haveData bool
	)

	pos := riffHeaderLen
	for pos+chunkHeaderLen <= len(wav) {
		id := string(wav[pos : pos+chunkIDLen])
		size := binary.LittleEndian.Uint32(wav[pos+chunkIDLen : pos+chunkHeaderLen])
		body := pos + chunkHeaderLen
		if uint64(body)+uint64(size) > uint64(len(wav)) {
			if id == chunkIDData && body <= len(wav) {
				dataSize = uint32(len(wav) - body)
				haveData = true
			}
			break
		}

		switch id {
		case chunkIDFmt:
			rate, err := parseByteRate(wav[body : body+int(size)])
			if err != nil {
				return 0, 0, err
			}
			byteRate = rate
			haveFmt = true
		case chunkIDData:
			dataSize = size
			haveData = true
		}

		pos = body + int(size)
		if size%chunkPaddingAlign == 1 {
			pos++
		}
	}

	if !haveFmt {
		return 0, 0, fmt.Errorf("missing fmt chunk")
	}
	if !haveData {
		return 0, 0, fmt.Errorf("missing data chunk")
	}
	return byteRate, dataSize, nil
}

func parseByteRate(seg []byte) (uint32, error) {
	if len(seg) < minFmtChunkLen {
		return 0, fmt.Errorf("fmt chunk shorter than 16 bytes: %d", len(seg))
	}
	byteRate := binary.LittleEndian.Uint32(seg[fmtByteRateOffset : fmtByteRateOffset+4])
	if byteRate != 0 {
		return byteRate, nil
	}
	channels := binary.LittleEndian.Uint16(seg[fmtChannelsOffset : fmtChannelsOffset+2])
	sampleRate := binary.LittleEndian.Uint32(seg[fmtSampleRateOff : fmtSampleRateOff+4])
	bitsPerSample := binary.LittleEndian.Uint16(seg[fmtBitsPerSampleOff : fmtBitsPerSampleOff+2])
	derived := sampleRate * uint32(channels) * uint32(bitsPerSample) / bitsPerByte
	if derived == 0 {
		return 0, fmt.Errorf("fmt chunk yields zero byteRate")
	}
	return derived, nil
}
