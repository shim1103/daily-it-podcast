package application_test

import (
	"context"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.ItemSource = (*fakeItemSource)(nil)

type fakeItemSource struct {
	calls []time.Time
	items []models.SourceItem
	err   error
}

func (f *fakeItemSource) List(_ context.Context, since time.Time) ([]models.SourceItem, error) {
	f.calls = append(f.calls, since)
	if f.err != nil {
		return nil, f.err
	}
	if f.items == nil {
		return []models.SourceItem{}, nil
	}
	return f.items, nil
}

var _ port.EpisodeWriter = (*fakeEpisodeWriter)(nil)

type fakeEpisodeWriter struct {
	calls      int
	episodeID  string
	manuscript []byte
	audio      models.SpeechAudio
	err        error
}

func (f *fakeEpisodeWriter) Write(_ context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	f.calls++
	f.episodeID = episodeID
	f.manuscript = manuscript
	f.audio = audio
	return f.err
}
