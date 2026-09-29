package port

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// EpisodeWriter は episode の原稿と音声を所定の object 空間へ書く。
// vendor HTTP・storage 固有 id・MIME は Infrastructure に閉じる。
// Application Gate（writeepisode）も本 IF を満たし、検査後に raw Adapter へ委譲してよい。
//
// @require episodeID は非空。audio.Content は非空 MP3 bytes。
// @ensure 成功時、所定空間直下に {episodeID}.json と {episodeID}.mp3 がある。途中失敗は成功にしない。
// @invariant storage 固有 id・MIME・vendor 型を露出しない。method は Write のみ。
type EpisodeWriter interface {
	Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error
}
