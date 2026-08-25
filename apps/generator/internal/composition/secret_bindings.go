package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport/processenv"
)

type secretBindings map[secrettransport.SecretRef]string

var _ secrettransport.BindingResolver = secretBindings(nil)

var (
	getXAPIKeySecret          = secrettransport.NewSecretRef()
	twitterIOAPIKeySecret     = secrettransport.NewSecretRef()
	geminiAPIKeySecret        = secrettransport.NewSecretRef()
	cursorAPIKeySecret        = secrettransport.NewSecretRef()
	googleOAuthClientIDSecret = secrettransport.NewSecretRef()
	googleOAuthClientSecret   = secrettransport.NewSecretRef()
	googleOAuthRefreshToken   = secrettransport.NewSecretRef()
	driveFolderIDSecret       = secrettransport.NewSecretRef()

	generatorSecretBindings = newSecretBindings()
	cursorCommandRuntime    = newCursorCommandRuntimeBinding()
)

type cursorCommandRuntimeBinding struct {
	apiKey                secrettransport.SecretRef
	inheritedEnvNameAllow [3]string
}

func newSecretBindings() secretBindings {
	return secretBindings{
		getXAPIKeySecret:          secretnames.GetXAPIKeyName,
		twitterIOAPIKeySecret:     secretnames.TwitterIOAPIKeyName,
		geminiAPIKeySecret:        secretnames.GeminiAPIKeyName,
		cursorAPIKeySecret:        secretnames.CursorAPIKeyName,
		googleOAuthClientIDSecret: secretnames.GoogleOAuthClientIDName,
		googleOAuthClientSecret:   secretnames.GoogleOAuthClientSecretName,
		googleOAuthRefreshToken:   secretnames.GoogleOAuthRefreshTokenName,
		driveFolderIDSecret:       secretnames.DriveFolderIDName,
	}
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

func (bindings secretBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := bindings[ref]
	return name, ok
}

// processenvSecretTransportClient は process environment 実装の secrettransport.Client を返す。
//
// @ensure 戻りは generatorSecretBindings で解決する secrettransport.Client。
func processenvSecretTransportClient() secrettransport.Client {
	return processenv.NewClient(generatorSecretBindings, nil, nil)
}
