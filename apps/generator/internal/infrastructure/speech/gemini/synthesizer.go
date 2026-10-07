package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
)

var _ port.SpeechSynthesizer = (*SpeechSynthesizer)(nil)

type SpeechSynthesizer struct {
	client         *http.Client
	apiKey         string
	tier           Tier
	retry          port.RetryReporter
	backoffSleepFn func(time.Duration) // why: test の並列実行と共存するため package global に置かない
	lastCallAt     time.Time
	nowFn          func() time.Time
	// why: 待機系パラメータは rate 計測 test から注入で差し替えるため field にする（Decision 2026-09-03T14-46-00）。
	callGap          time.Duration
	retryBackoffBase time.Duration
	retryBackoffMax  time.Duration
}

// SynthesizeAll は texts を順に朗読音声へ変換し、セグメント単位の WAV 列（結合しない）を返す。
//
// @require texts の各要素は trim 後に非空。
// @ensure 成功時は len(texts) と同数の非空・最小尺 WAV を返す。
// @ensure 失敗時もそれまでに合成できた分の audios（部分成功）を err と併せて返す。
// @ensure 呼び出し全体で Gemini 呼び出し合計を Tier ごとの上限（SynthesizeBudget / SynthesizeBudgetPaid）以内に抑える。1 セグメントは min(MaxAttempts, 残予算) 回まで。上限へ達した後のセグメントは即 error。
// @ensure 取得元が当面使えない失敗（認証断・利用枠喪失・回復の明示が無い 429・再試行の使い切り）は Tier を問わず port.ErrSourceExhausted を wrap して返す。それ以外（4xx・呼び出し上限超過など）は wrap せずそのまま返す（Decision generator-api-failure-retry-or-fallback）。
func (s *SpeechSynthesizer) SynthesizeAll(ctx context.Context, texts []string) ([]models.SpeechAudio, error) {
	if s == nil || s.client == nil {
		return nil, infraErr("synthesize", fmt.Errorf("client is nil"))
	}

	budget := s.synthesizeBudget()
	audios := make([]models.SpeechAudio, 0, len(texts))
	callsSpent := 0
	for i, text := range texts {
		remaining := budget - callsSpent
		if remaining <= 0 {
			return audios, budgetExhaustedError(i, len(texts), callsSpent, budget)
		}
		audio, used, err := s.synthesizeOne(ctx, text, min(MaxAttempts, remaining))
		callsSpent += used
		if err != nil {
			return audios, err
		}
		audios = append(audios, audio)
	}
	return audios, nil
}

func budgetExhaustedError(segmentIndex, segmentCount, spent, budget int) error {
	return infraErr("synthesize_budget", fmt.Errorf(
		"gemini call budget exhausted at segment %d/%d: spent %d of %d", segmentIndex+1, segmentCount, spent, budget))
}

func (s *SpeechSynthesizer) synthesizeBudget() int {
	if s.tier == TierPaid {
		return SynthesizeBudgetPaid
	}
	return SynthesizeBudget
}

// why: adaptererror.Error は全 infra 共通型なので、Source も見ないと別 Adapter の同名 Op（"http_status" 等）が偶然一致しうる。
func sameGeminiOp(prev, cur error) bool {
	if prev == nil || cur == nil {
		return false
	}
	var prevErr, curErr *adaptererror.Error
	if !errors.As(prev, &prevErr) || !errors.As(cur, &curErr) {
		return false
	}
	return prevErr.Source == curErr.Source && prevErr.Op == curErr.Op
}
