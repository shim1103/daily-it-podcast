package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

// newCursorTextWriter は Cursor CLI TextWriter Adapter を組み立てる。
// secret 値（cfg.APIKey.Reveal()）を processenv factory closure に閉じ、
// cursorcli には inject env 名を受け取る factory のみを渡す。
// cursorcli は secret 値・lookupEnv・runtime 実装を知らない。
//
// @require cfg.APIKey は validation 済み。
// @ensure 戻りは port.TextWriter。cursorcli は secret 値を保持しない。
func newCursorTextWriter(cfg config.CursorConfig) port.TextWriter {
	factory := processenv.NewSecretEnvLauncherFactory(
		cfg.APIKey.Reveal(),
		sharedLookupEnv(),
		[]string{
			config.GetXAPIKeyEnv,
			config.GeminiAPIKeyEnv,
			config.GoogleOAuthClientSecretEnv,
			config.GoogleOAuthRefreshTokenEnv,
		},
	)
	return cursorcli.NewTextWriter(factory)
}
