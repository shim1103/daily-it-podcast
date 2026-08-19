package port

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// EpisodeWriter は episode の原稿と音声を所定フォルダへ書く。
// vendor HTTP・Drive file id・MIME は Infrastructure に閉じる。
//
// @require episodeID は非空。manuscript は manuscript.schema.json に適合し、中の episodeId は episodeID と一致する。audio.Content は非空 WAV bytes。
// @ensure 成功時、所定フォルダ直下に {episodeID}.json と {episodeID}.wav がある。途中失敗は成功にしない。
// @invariant Drive file id・MIME・vendor 型を露出しない。method は Write のみ。mp3 を書かない。
type EpisodeWriter interface {
	Write(ctx context.Context, episodeID string, manuscript []byte, audio models.SpeechAudio) error
}
