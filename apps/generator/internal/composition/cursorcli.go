package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

// NewCursorTextWriter は Cursor CLI TextWriter Adapter を組み立てる。
//
// @ensure 戻りは port.TextWriter。
// @ensure production path の launcher は process-env 実装であり、AgentSecrets EnvWrapper を置換しない。
func NewCursorTextWriter() port.TextWriter {
	launcher := processenv.NewLauncher(
		generatorSecretBindings,
		cursorCommandRuntime.apiKey,
		cursorCommandRuntime.inheritedEnvNameAllow[:],
	)
	return cursorcli.NewTextWriter(launcher)
}
