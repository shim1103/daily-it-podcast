package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
)

// NewGoogleDriveWriteEpisode は Google Drive を保存先とする UseCase を組み立てる。
//
// @ensure 戻りは validation 後にだけ raw adapter を呼ぶ。
func NewGoogleDriveWriteEpisode() *application.WriteEpisode {
	client := processenvSecretTransportClient()
	tokens := oauth.NewTokenSource(client, googleOAuthClientIDSecret, googleOAuthClientSecret, googleOAuthRefreshToken)
	rawWriter := gdrive.NewRawEpisodeWriter(client, tokens, driveFolderIDSecret)
	return application.NewWriteEpisode(rawWriter)
}
