package application

import (
	"context"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

type stubItemSource struct{}

func (stubItemSource) List(context.Context, time.Time) ([]models.SourceItem, error) {
	return nil, nil
}

type stubTextWriter struct{}

func (stubTextWriter) Write(context.Context, string) (string, error) {
	return "断片", nil
}

type stubSpeechSynthesizer struct{}

func (stubSpeechSynthesizer) Synthesize(context.Context, string) (models.SpeechAudio, error) {
	return models.SpeechAudio{Content: []byte("RIFF")}, nil
}

type stubEpisodeWriter struct{}

func (stubEpisodeWriter) Write(context.Context, string, []byte, models.SpeechAudio) error {
	return nil
}

var (
	_ port.ItemSource        = stubItemSource{}
	_ port.TextWriter        = stubTextWriter{}
	_ port.SpeechSynthesizer = stubSpeechSynthesizer{}
	_ port.EpisodeWriter     = stubEpisodeWriter{}
)

func TestWavDurationSec_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("wavDurationSec: want panic")
		}
	}()
	_, _ = wavDurationSec([]byte{1, 2})
}

func TestConcatWAV_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("concatWAV: want panic")
		}
	}()
	_, _ = concatWAV([]byte{1, 2})
}

func TestManuscriptDraftFromWriterOutput_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("manuscriptDraftFromWriterOutput: want panic")
		}
	}()
	_, _ = manuscriptDraftFromWriterOutput("断片")
}

func TestProduceEpisodeRun_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	uc := NewProduceEpisode(
		NewFetchSourceItems(stubItemSource{}),
		stubTextWriter{},
		stubSpeechSynthesizer{},
		NewWriteEpisode(stubEpisodeWriter{}),
	)
	defer func() {
		if recover() == nil {
			t.Fatal("ProduceEpisode.Run: want panic")
		}
	}()
	_ = uc.Run(context.Background(), time.Unix(0, 0).UTC())
}
