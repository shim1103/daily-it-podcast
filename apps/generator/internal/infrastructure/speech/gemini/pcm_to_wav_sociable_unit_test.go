package gemini

import "testing"

func TestPCMToWAV_returnsNonEmptyWAV_whenMinimalMonoPCM(t *testing.T) {
	t.Parallel()

	// Given: 16-bit mono PCM
	pcm := minimalPCM()

	// When: WAV へ変換する
	got, err := pcmToWAV(pcm)

	// Then: 非空 RIFF/WAVE header がある
	if err != nil {
		t.Fatalf("pcmToWAV: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("output is empty")
	}
	if !isWAV(got) {
		t.Fatalf("output head = % x, want RIFF/WAVE", got[:min(12, len(got))])
	}
}

func TestPCMToWAV_returnsError_whenPCMEmpty(t *testing.T) {
	t.Parallel()

	// Given: 空 PCM
	pcm := []byte{}

	// When: WAV へ変換する
	_, err := pcmToWAV(pcm)

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPCMToWAV_returnsError_whenPCMLengthOdd(t *testing.T) {
	t.Parallel()

	// Given: 奇数 byte の PCM
	pcm := []byte{0x00, 0x01, 0x02}

	// When: WAV へ変換する
	_, err := pcmToWAV(pcm)

	// Then: error が返る
	if err == nil {
		t.Fatal("expected error")
	}
}
