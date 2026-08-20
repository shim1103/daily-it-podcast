package composition

import (
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
)

// NewGoogleDriveEpisodeWriter は Google Drive 書込 Adapter を組み立てる。
//
// @require tokens != nil
// @ensure 戻りは port.EpisodeWriter。
func NewGoogleDriveEpisodeWriter(tokens gdrive.TokenSource) port.EpisodeWriter {
	return gdrive.NewEpisodeWriter(&agentsecrets.Client{}, tokens)
}
