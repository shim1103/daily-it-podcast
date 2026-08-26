package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
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
)

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

func (bindings secretBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := bindings[ref]
	return name, ok
}
