package build_test

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
)

// scaffold: stub の零値契約だけを固定する。変換挙動の TDD は C（Issue）側。
func TestEncodeWAVToMP3_stubReturnsNilNil(t *testing.T) {
	t.Parallel()

	// Given: 任意の WAV バイト（内容は問わない）
	wav := []byte("not-a-real-wav")

	// When: stub を呼ぶ
	got, err := build.EncodeWAVToMP3(wav)

	// Then: 零値（nil, nil）
	if err != nil {
		t.Fatalf("EncodeWAVToMP3: err = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("EncodeWAVToMP3: got = %v, want nil", got)
	}
}
