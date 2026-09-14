package r2_test

import (
	"context"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

// scaffold: stub の零値契約だけを固定する。Put 挙動の TDD は C（Issue）側。
func TestWrite_stubReturnsNil(t *testing.T) {
	t.Parallel()

	// Given: stub Writer と任意の入力
	w := r2.NewEpisodeWriter()

	// When: stub を呼ぶ
	err := w.Write(context.Background(), "ep-1", []byte(`{}`), models.SpeechAudio{Content: []byte("mp3")})

	// Then: nil
	if err != nil {
		t.Fatalf("Write: err = %v, want nil", err)
	}
}
