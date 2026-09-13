package ffmpeg_test

import (
	"context"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/audio/ffmpeg"
)

// scaffold: stub の零値契約だけを固定する。変換挙動の TDD は C（Issue）側。
func TestEncodeWAVToMP3_stubReturnsNilNil(t *testing.T) {
	t.Parallel()

	// Given: 任意の WAV バイト（内容は問わない）と stub Encoder
	enc := ffmpeg.NewEncoder(nil, nil)
	wav := []byte("not-a-real-wav")

	// When: stub を呼ぶ
	got, err := enc.EncodeWAVToMP3(context.Background(), wav)

	// Then: 零値（nil, nil）
	if err != nil {
		t.Fatalf("EncodeWAVToMP3: err = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("EncodeWAVToMP3: got = %v, want nil", got)
	}
}
