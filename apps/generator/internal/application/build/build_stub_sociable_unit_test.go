package build_test

import (
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/build"
)

func TestManuscriptDraftFromWriterOutput_panicsAsContractStub(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("ManuscriptDraftFromWriterOutput: want panic")
		}
	}()
	_, _ = build.ManuscriptDraftFromWriterOutput("断片")
}
