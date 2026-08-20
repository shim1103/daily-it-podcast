package composition_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/composition"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

type fakeTokenSource struct {
	calls int
}

func (s *fakeTokenSource) Token(context.Context) (string, error) {
	s.calls++
	return "unused", nil
}

func TestNewGoogleDriveWriteEpisode_validatesBeforeDriveAccess(t *testing.T) {
	// Given: composition factory で組み立てた writer
	tokens := &fakeTokenSource{}
	useCase := composition.NewGoogleDriveWriteEpisode(tokens)

	// When: 空 episodeID で UseCase を実行する
	err := useCase.Run(context.Background(), "", []byte(`{}`), models.SpeechAudio{Content: []byte("RIFFWAV")})

	// Then: Application の Domain Error。Drive access は発生しない
	var emptyID *domainerrors.EmptyEpisodeID
	if !errors.As(err, &emptyID) {
		t.Fatalf("error type %T (%v), want *errors.EmptyEpisodeID", err, err)
	}
	if tokens.calls != 0 {
		t.Fatalf("Token calls = %d, want 0", tokens.calls)
	}
}
