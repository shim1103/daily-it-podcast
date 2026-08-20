package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
)

// NewGoogleDriveWriteEpisode は Google Drive を保存先とする UseCase を組み立てる。
//
// @require tokens != nil
// @ensure 戻りは validation 後にだけ raw adapter を呼ぶ。
func NewGoogleDriveWriteEpisode(tokens gdrive.TokenSource) *application.WriteEpisode {
	rawWriter := gdrive.NewRawEpisodeWriter(&agentsecrets.Client{}, tokens)
	return application.NewWriteEpisode(rawWriter)
}
