package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

// NewCursorTextWriter は Cursor CLI TextWriter Adapter を組み立てる。
//
// @ensure 戻りは port.TextWriter。
func NewCursorTextWriter() port.TextWriter {
	return cursorcli.NewTextWriter()
}
