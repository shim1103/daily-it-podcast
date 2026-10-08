package gemini

import (
	"encoding/binary"
	"fmt"
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
	if len(wav) < 12 {
		return fmt.Errorf("wav is shorter than RIFF header: %d bytes", len(wav))
	}
	if string(wav[0:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
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

	pos := 12
	for pos+8 <= len(wav) {
		id := string(wav[pos : pos+4])
		size := binary.LittleEndian.Uint32(wav[pos+4 : pos+8])
		body := pos + 8
		if uint64(body)+uint64(size) > uint64(len(wav)) {
			if id == "data" && body <= len(wav) {
				dataSize = uint32(len(wav) - body)
				haveData = true
			}
			break
		}

		switch id {
		case "fmt ":
			rate, err := parseByteRate(wav[body : body+int(size)])
			if err != nil {
				return 0, 0, err
			}
			byteRate = rate
			haveFmt = true
		case "data":
			dataSize = size
			haveData = true
		}

		pos = body + int(size)
		if size%2 == 1 {
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
	if len(seg) < 16 {
		return 0, fmt.Errorf("fmt chunk shorter than 16 bytes: %d", len(seg))
	}
	byteRate := binary.LittleEndian.Uint32(seg[8:12])
	if byteRate != 0 {
		return byteRate, nil
	}
	channels := binary.LittleEndian.Uint16(seg[2:4])
	sampleRate := binary.LittleEndian.Uint32(seg[4:8])
	bitsPerSample := binary.LittleEndian.Uint16(seg[14:16])
	derived := sampleRate * uint32(channels) * uint32(bitsPerSample) / 8
	if derived == 0 {
		return 0, fmt.Errorf("fmt chunk yields zero byteRate")
	}
	return derived, nil
}
