package build_test

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
)

func TestWavDurationSec_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("WavDurationSec: want panic")
		}
	}()
	_, _ = build.WavDurationSec([]byte{1, 2})
}

func TestConcatWAV_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("ConcatWAV: want panic")
		}
	}()
	_, _ = build.ConcatWAV([]byte{1, 2})
}

func TestManuscriptDraftFromWriterOutput_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("ManuscriptDraftFromWriterOutput: want panic")
		}
	}()
	_, _ = build.ManuscriptDraftFromWriterOutput("断片")
}
