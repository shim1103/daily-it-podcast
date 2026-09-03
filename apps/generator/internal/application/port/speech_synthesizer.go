package port

import (
	"context"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// SpeechSynthesizer は朗読本文の列を音声へ変換する。
// vendor HTTP・PCM→WAV wrap・演出文は Infrastructure に閉じる。
//
// @require texts の各要素は trim 後に非空。朗読本文のみ。演出・voice は渡さない。
// @ensure 成功時は texts と同数の非空 WAV bytes 列（セグメント単位。結合しない）。同一 text でも byte 列の一致は約束しない。
// @ensure 1 回の呼び出し全体で vendor 呼び出し合計を無料枠内へ抑える（retry 予算は Adapter が束ねて管理する）。
// @invariant vendor 固有型・PCM・path・voice 名を露出しない。method は SynthesizeAll のみ。
type SpeechSynthesizer interface {
	SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error)
}
