package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
)

// newGoogleDriveWriteEpisode は Google Drive を保存先とする UseCase を組み立てる。
//
// @require httpClient != nil。cfg は検証済み。
// @ensure 戻りは validation 後にだけ raw adapter を呼ぶ。
func newGoogleDriveWriteEpisode(httpClient *http.Client, cfg config.DriveConfig) *application.WriteEpisode {
	tokens := oauth.NewTokenSource(
		httpClient,
		cfg.GoogleOAuthClientID,
		cfg.GoogleOAuthClientSecret.Reveal(),
		cfg.GoogleOAuthRefreshToken.Reveal(),
	)
	rawWriter := gdrive.NewRawEpisodeWriter(httpClient, tokens, cfg.FolderID)
	return application.NewWriteEpisode(rawWriter)
}

// newGoogleDriveCompletedEpisodeLookup は Google Drive を照会先とする Port 実装を返す。
//
// @require httpClient != nil。cfg は検証済み。
// @ensure 戻りは非 nil の port.CompletedEpisodeLookup。
func newGoogleDriveCompletedEpisodeLookup(httpClient *http.Client, cfg config.DriveConfig) port.CompletedEpisodeLookup {
	tokens := oauth.NewTokenSource(
		httpClient,
		cfg.GoogleOAuthClientID,
		cfg.GoogleOAuthClientSecret.Reveal(),
		cfg.GoogleOAuthRefreshToken.Reveal(),
	)
	return gdrive.NewCompletedEpisodeLookup(httpClient, tokens, cfg.FolderID)
}
