package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/writeepisode"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

// newR2WriteEpisode は R2 raw Adapter を Gate（writeepisode）で包む結線口である。
//
// @require httpClient != nil。cfg は config.Load で検証済みの R2Config（Config.R2）。
// @ensure 戻りは port.EpisodeWriter（検証後にだけ raw を呼ぶ）。
func newR2WriteEpisode(httpClient *http.Client, cfg config.R2Config) port.EpisodeWriter {
	rawWriter := r2.NewEpisodeWriter(
		httpClient,
		cfg.AccessKeyID.Reveal(),
		cfg.SecretAccessKey.Reveal(),
		cfg.AccountID,
		cfg.Bucket,
	)
	return writeepisode.NewWriteEpisode(rawWriter)
}

// newR2CompletedEpisodeLookup は R2 を照会先とする Port 実装を組み立てる結線口である。
// newProduceEpisode は本結線口を正本として使う（列 6 で Writer と同着で完全置換済み）。
//
// @require httpClient != nil。cfg は config.Load で検証済みの R2Config（Config.R2）。
// @ensure 戻りは非 nil の port.CompletedEpisodeLookup。
func newR2CompletedEpisodeLookup(httpClient *http.Client, cfg config.R2Config) port.CompletedEpisodeLookup {
	return r2.NewCompletedEpisodeLookup(
		httpClient,
		cfg.AccessKeyID.Reveal(),
		cfg.SecretAccessKey.Reveal(),
		cfg.AccountID,
		cfg.Bucket,
	)
}
