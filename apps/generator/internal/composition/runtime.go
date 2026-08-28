package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	commandlaunchprocessenv "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
)

var cursorCommandRuntime = newCursorCommandRuntimeBinding()

type cursorCommandRuntimeBinding struct {
	apiKey                secrettransport.SecretRef
	inheritedEnvNameAllow [3]string
}

func newCursorCommandRuntimeBinding() cursorCommandRuntimeBinding {
	return cursorCommandRuntimeBinding{
		apiKey: cursorAPIKeySecret,
		inheritedEnvNameAllow: [3]string{
			"PATH",
			"HOME",
			"TMPDIR",
		},
	}
}

// CursorCommandInheritedEnvNameAllow は Cursor command child へ継承を許す env 名の allowlist を返す。
//
// @ensure 戻りは Composition 所有の allowlist のコピーであり、呼び出し側が変更しても SSoT を壊さない。
func CursorCommandInheritedEnvNameAllow() []string {
	return append([]string(nil), cursorCommandRuntime.inheritedEnvNameAllow[:]...)
}

// processenvSecretTransportClient は process environment 実装の secrettransport.Client を返す。
//
// @ensure 戻りは generatorSecretBindings で解決する secrettransport.Client。
func processenvSecretTransportClient() secrettransport.Client {
	return processenv.NewClient(generatorSecretBindings, nil, nil)
}

// processenvCursorCommandLauncher は production 既定の commandlaunch.Launcher を返す。
func processenvCursorCommandLauncher() commandlaunch.Launcher {
	return commandlaunchprocessenv.NewLauncher(
		generatorSecretBindings,
		cursorCommandRuntime.apiKey,
		cursorCommandRuntime.inheritedEnvNameAllow[:],
		nil,
	)
}
