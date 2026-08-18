package port

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// SpeechSynthesizer は本文 1 件を音声 file へ変換する。
// vendor HTTP・PCM→WAV wrap・演出文は Infrastructure に閉じる。
//
// @require text は trim 後に非空。朗読本文のみ。演出・voice は渡さない。
// @ensure 成功時は非空の WAV bytes。同一 text でも byte 列の一致は約束しない。
// @invariant vendor 固有型・PCM・path・voice 名を露出しない。method は Synthesize のみ。
type SpeechSynthesizer interface {
	Synthesize(ctx context.Context, text string) (models.SpeechAudio, error)
}
