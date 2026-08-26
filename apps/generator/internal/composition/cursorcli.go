package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

// NewCursorTextWriter は Cursor CLI TextWriter Adapter を組み立てる。
//
// @ensure 戻りは port.TextWriter。
// @ensure production path の launcher は process-env 実装であり、AgentSecrets launcher を置換しない。
func NewCursorTextWriter() port.TextWriter {
	return cursorcli.NewTextWriter(processenvCursorCommandLauncher())
}

// NewCursorTextWriterLocal は local 選択時に AgentSecrets launcher を結線した TextWriter を返す。
//
// @ensure 戻りは port.TextWriter。
// @ensure launcher は Composition 所有の Cursor 専用 project で注入範囲を閉じる。
func NewCursorTextWriterLocal() port.TextWriter {
	return cursorcli.NewTextWriter(agentsecretsCursorCommandLauncher())
}
