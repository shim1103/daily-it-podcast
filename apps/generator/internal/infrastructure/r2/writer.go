// Package r2 は Cloudflare R2（S3 互換）へ episode を書く Driven Adapter である。
package r2

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

var _ port.EpisodeWriter = (*EpisodeWriter)(nil)

// EpisodeWriter は R2 へ {episodeId}.json / {episodeId}.mp3 を put する。
// 本 stub の Write は零値成功のみ。S3 本実装は C（Issue）側。
type EpisodeWriter struct{}

// NewEpisodeWriter は R2 EpisodeWriter stub を返す。
//
// @ensure 戻りは非 nil の *EpisodeWriter（port.EpisodeWriter）。
func NewEpisodeWriter() *EpisodeWriter {
	return &EpisodeWriter{}
}

// Write は原稿と音声を所定 bucket へ書く。
//
// @ensure 現 stub は常に nil を返す。
func (w *EpisodeWriter) Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error {
	_ = ctx
	_ = episodeID
	_ = manuscript
	_ = audio
	_ = w
	return nil
}
