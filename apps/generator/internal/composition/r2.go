package composition

import (
	"net/http"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

// newR2WriteEpisode は R2 を保存先とする UseCase を組み立てる結線口である。
// 現行 newProduceEpisode は Drive を正本として結線する。本番切替は列 6。
//
// @require httpClient != nil。cfg は LoadR2 で検証済み。
// @ensure 戻りは validation 後にだけ raw adapter を呼ぶ。
func newR2WriteEpisode(httpClient *http.Client, cfg config.R2Config) *application.WriteEpisode {
	rawWriter := r2.NewEpisodeWriter(
		httpClient,
		r2.Credentials{
			AccessKeyID:     cfg.AccessKeyID.Reveal(),
			SecretAccessKey: cfg.SecretAccessKey.Reveal(),
		},
		r2.Endpoint{
			AccountID: cfg.AccountID,
			Bucket:    cfg.Bucket,
		},
	)
	return application.NewWriteEpisode(rawWriter)
}
