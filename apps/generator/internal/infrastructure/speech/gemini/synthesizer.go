package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

const geminiAPIKeyHeader = "x-goog-api-key"

var _ port.SpeechSynthesizer = (*SpeechSynthesizer)(nil)

// synthesizeBudget と maxAttempts の関係:
//   - MaxAttempts   : 1 セグメントが連続で消費してよい上限（暴走ガード）。
//   - SynthesizeBudget: 1 度の SynthesizeAll 全体で許す合計上限（RPD ガード）。
//     各セグメントは min(MaxAttempts, 残予算) 回まで。

type SpeechSynthesizer struct {
	client         *http.Client
	apiKey         string
	backoffSleepFn func(time.Duration) // why: test の並列実行と共存するため package global に置かない
	lastCallAt     time.Time
	nowFn          func() time.Time
	// why: 429 頻度と総所要のトレードオフを実測で詰めるため、待機系パラメータを field 化して
	//      rate 計測 test から注入で差し替える（Decision 2026-09-03T14-46-00）。
	//      既定 constructor は default* const を入れるので挙動は不変。MaxAttempts は注入対象外。
	callGap          time.Duration
	retryBackoffBase time.Duration
	retryBackoffMax  time.Duration
}

// SynthesizeAll は texts を順に朗読音声へ変換し、セグメント単位の WAV 列（結合しない）を返す。
// retry 予算・callGap・RPD quota は Adapter 定数 = vendor 固有制約であり、
// 「1 episode 分の TTS 呼び出し群」を束ねて管理するのは Adapter の責務。
//
// @require texts の各要素は trim 後に非空。朗読本文のみ。
// @ensure 成功時は len(texts) と同数の非空・最小尺 WAV を返す（結合しない）。
// @ensure 呼び出し全体で Gemini 呼び出し合計を SynthesizeBudget 回以内へ抑える。
//
//	1 セグメントは min(MaxAttempts, 残予算) 回まで。合計が SynthesizeBudget へ達したら以降のセグメントは即 error。
func (s *SpeechSynthesizer) SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("synthesize", fmt.Errorf("client is nil"))
	}

	audios := make([]models.SpeechAudio, 0, len(texts))
	callsSpent := 0
	for i, text := range texts {
		remaining := SynthesizeBudget - callsSpent
		if remaining <= 0 {
			// why: ここへ来る前に必ず 1 セグメント以上を消費している。合計予算が尽きた。
			return nil, infraErr("synthesize_budget", fmt.Errorf(
				"gemini call budget exhausted at segment %d/%d: spent %d of %d", i+1, len(texts), callsSpent, SynthesizeBudget))
		}
		maxAttempts := MaxAttempts
		if remaining < maxAttempts {
			maxAttempts = remaining
		}
		audio, used, err := s.synthesizeOne(ctx, text, maxAttempts)
		callsSpent += used
		if err != nil {
			return nil, err
		}
		audios = append(audios, audio)
	}
	return audios, nil
}

// sameGeminiOp は 2 つの error がともに *gemini.Error で Op 文字列が一致するかを返す。
// 片方でも *gemini.Error でなければ false。
func sameGeminiOp(prev, cur error) bool {
	if prev == nil || cur == nil {
		return false
	}
	var prevErr, curErr *Error
	if !errors.As(prev, &prevErr) || !errors.As(cur, &curErr) {
		return false
	}
	return prevErr.Op == curErr.Op
}
