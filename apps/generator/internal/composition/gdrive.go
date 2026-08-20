package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
)

// NewGoogleDriveEpisodeWriter は Google Drive 書込 Adapter を組み立てる。
//
// @ensure 戻りは port.EpisodeWriter。
func NewGoogleDriveEpisodeWriter() port.EpisodeWriter {
	client := &agentsecrets.Client{}
	return gdrive.NewEpisodeWriter(client, oauth.NewTokenSource(client))
}
