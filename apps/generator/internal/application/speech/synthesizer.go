// Package speech は音声合成の UseCase を提供する。
// source を順に試し、枯渇したら部分成功を保持したまま残りの texts だけを次 source へ渡して継続する。
package speech

import (
	"context"
	"errors"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

var _ port.SpeechSynthesizer = (*SpeechSynthesizer)(nil)

// why: event 名の typo は compile で捕まらないので定数化する。
const fallbackEventTTSSourceSwitched = "tts_source_switched"

// SpeechSynthesizer は sources を先頭から順に試し音声を合成する UseCase である。
type SpeechSynthesizer struct {
	sources  []port.SpeechSynthesizer
	fallback port.FallbackReporter
}

// NewSpeechSynthesizer は sources と、切り替え発生時の観測面を束ねる。
//
// @require len(sources) > 0 かつ全要素が非 nil。fallback != nil。
// @ensure 戻りは *SpeechSynthesizer（port.SpeechSynthesizer を満たす）。
func NewSpeechSynthesizer(sources []port.SpeechSynthesizer, fallback port.FallbackReporter) *SpeechSynthesizer {
	return &SpeechSynthesizer{sources: sources, fallback: fallback}
}

// SynthesizeAll は sources を先頭から順に試す。
//
// @require texts の各要素は trim 後に非空。各 source.SynthesizeAll へそのまま渡す。
// @ensure ある source が残り texts 全件の合成に成功したら、それまでに蓄積した分と合わせて全 audios を返す。
// @ensure source の error が errors.Is(err, port.ErrSourceExhausted)==true のとき、その source が返した
//
//	部分成功分（got）を蓄積し、fallback.Fallback(fallbackEventTTSSourceSwitched) を呼んでから
//	残りの texts（remaining[len(got):]）だけを次 source へ渡す。
//
// @ensure source の error が port.ErrSourceExhausted でないとき、その error をそのまま返す。
//
//	蓄積した部分成功分は破棄する（枯渇以外は全体失敗として扱う）。
//
// @ensure 全 source を使い切ったら最後の source の error を返す。
// @invariant sources は互いに独立（直交）している。ある source の内部状態・retry 予算・quota は
//
//	他 source に伝播しない。fallback 判断は port.ErrSourceExhausted の有無だけで行い、vendor を問わない。
func (s *SpeechSynthesizer) SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error) {
	remaining := texts
	accumulated := make([]models.SpeechAudio, 0, len(texts))
	var lastErr error

	for _, source := range s.sources {
		got, err := source.SynthesizeAll(ctx, remaining)
		if err == nil {
			return append(accumulated, got...), nil
		}
		lastErr = err

		if !errors.Is(err, port.ErrSourceExhausted) {
			return nil, err
		}

		accumulated = append(accumulated, got...)
		remaining = remaining[len(got):]
		s.fallback.Fallback(fallbackEventTTSSourceSwitched)
	}

	return nil, lastErr
}
