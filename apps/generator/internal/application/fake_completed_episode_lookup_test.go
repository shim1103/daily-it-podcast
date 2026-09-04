package application_test

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.CompletedEpisodeLookup = (*fakeCompletedEpisodeLookup)(nil)

// fakeCompletedEpisodeLookup は CompletedEpisodeLookup の Fake。有無と error を制御し、照会 date を記録する。
type fakeCompletedEpisodeLookup struct {
	has      bool
	err      error
	calls    int
	lastDate string
}

func (f *fakeCompletedEpisodeLookup) HasPair(_ context.Context, date string) (bool, error) {
	f.calls++
	f.lastDate = date
	if f.err != nil {
		return false, f.err
	}
	return f.has, nil
}
