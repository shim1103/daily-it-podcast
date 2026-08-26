package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	commandlaunchagentsecrets "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/agentsecrets"
	commandlaunchprocessenv "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/agentsecrets"
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

// CursorCommandProjectDir は Cursor 専用 AgentSecrets project の設定 dir を返す。
//
// @ensure 戻りは Composition が所有する Cursor 専用 project の path（home 配下規約は commandlaunch/agentsecrets）。
// @ensure HOME が絶対 path のとき戻りも絶対 path。
func CursorCommandProjectDir() string {
	return commandlaunchagentsecrets.DefaultProjectDir(commandlaunchagentsecrets.CursorProjectName)
}

// processenvSecretTransportClient は process environment 実装の secrettransport.Client を返す。
//
// @ensure 戻りは generatorSecretBindings で解決する secrettransport.Client。
func processenvSecretTransportClient() secrettransport.Client {
	return processenv.NewClient(generatorSecretBindings, nil, nil)
}

// agentsecretsSecretTransportClient は local 向け AgentSecrets proxy 実装の secrettransport.Client を返す。
//
// @ensure 戻りは generatorSecretBindings で解決する secrettransport.Client。
func agentsecretsSecretTransportClient() secrettransport.Client {
	return agentsecrets.NewClient(generatorSecretBindings, nil, "")
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

// agentsecretsCursorCommandLauncher は local 向け AgentSecrets commandlaunch.Launcher を返す。
func agentsecretsCursorCommandLauncher() commandlaunch.Launcher {
	return commandlaunchagentsecrets.NewLauncher(
		CursorCommandProjectDir(),
		cursorCommandSecretKeys(),
		cursorCommandRuntime.inheritedEnvNameAllow[:],
		nil,
		nil,
	)
}

// cursorCommandSecretKeys は Cursor command が依存する秘密名の宣言を返す。
func cursorCommandSecretKeys() []string {
	name, ok := generatorSecretBindings.ResolveSecret(cursorCommandRuntime.apiKey)
	if !ok || name == "" {
		return nil
	}
	return []string{name}
}
