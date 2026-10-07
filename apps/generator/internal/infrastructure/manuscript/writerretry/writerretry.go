// Package writerretry は、原稿 TextWriter Adapter（geminiapi・cursorapi）が共有する、
// 再試行と fallback 判定の helper を提供する。
// 取得元ごとに違う request の組み立てと response の解釈は、各 Adapter が持つ。
package writerretry

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

const (
	// MaxRateLimitedAttempts は Retry-After 付き 429 を待って再試行する最大試行数。無限 retry を防ぐ。
	MaxRateLimitedAttempts = 4
	// MaxDraftAttempts は ManuscriptDraft 検証失敗（invalid-draft）時の Write 内部 retry 上限。
	MaxDraftAttempts = 5
	// MaxRetryAfter は Retry-After 由来の待ち時間の上限。
	// why: 異常値で run を止めない。run 全体の上限は GHA job / process cancel に委ねる。
	MaxRetryAfter = 30 * time.Second
)

type Kind int

const (
	// None は再試行せず、error をそのまま返す（bug）。
	None Kind = iota
	// Once は +1 即再試行を 1 回だけ行い、使い切りは fallback へ渡す。
	Once
	// RateLimited は Retry-After で回復が明示された 429。MaxRateLimitedAttempts まで待って再試行し、使い切りは fallback へ渡す。
	RateLimited
	// Fallback は再試行せず fallback へ渡す。
	Fallback
)

// warn: zero value は None（再試行しない）。成功と再試行しない失敗は Policy{} で返す。
type Policy struct {
	Kind Kind
	Wait time.Duration
}

// why: 429 は body の文言・code が provider 依存で変わりうるので読まず、標準 header の Retry-After だけを
// 回復の明示とする（Decision generator-genai-api-failure-retry-or-fallback）。
func ClassifyStatus(status int, header http.Header) Policy {
	switch {
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		return Policy{Kind: Fallback}
	case status == http.StatusTooManyRequests:
		if wait := retryAfter(header); wait > 0 {
			return Policy{Kind: RateLimited, Wait: wait}
		}
		return Policy{Kind: Fallback}
	case status >= http.StatusInternalServerError:
		return Policy{Kind: Once}
	default:
		return Policy{Kind: None}
	}
}

// why: delta-seconds 形式だけを回復の明示とする。HTTP-date 形式は解釈せず「明示なし」にする
// （YAGNI。Gemini と Cursor の 429 は delta-seconds で返る想定）。
func retryAfter(header http.Header) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(header.Get("Retry-After")))
	if err != nil || secs <= 0 {
		return 0
	}
	return min(time.Duration(secs)*time.Second, MaxRetryAfter)
}

// WrapForFallback は、kind が bug（None）以外のとき err を fallback の番兵で wrap して返す。
func WrapForFallback(kind Kind, err error) error {
	if kind == None {
		return err
	}
	return PassToFallback(err)
}

// PassToFallback は err を、別の取得元へ切り替えてよい合図の番兵で wrap する。
func PassToFallback(err error) error {
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, err)
}

// Delivery は Run が再試行を外へ出す出口（通知と待ち）を持つ。
type Delivery struct {
	// Step は再試行の通知（Retry）へ載せる処理名。
	Step  string
	Retry port.RetryReporter
	Sleep func(context.Context, time.Duration)
}

// Run は fetch を、返された Policy に従って再試行する。再試行の方針はこの関数だけが持ち、fetch は分類だけを返す。
func Run[T any](ctx context.Context, delivery Delivery, fetch func(context.Context) (T, Policy, error)) (T, error) {
	var zero T
	retriedOnce := false
	rateLimitedAttempts := 0
	for {
		value, policy, err := fetch(ctx)
		if err == nil {
			return value, nil
		}
		switch policy.Kind {
		case Once:
			if retriedOnce {
				return zero, WrapForFallback(policy.Kind, err)
			}
			retriedOnce = true
		case RateLimited:
			rateLimitedAttempts++
			if rateLimitedAttempts >= MaxRateLimitedAttempts {
				return zero, WrapForFallback(policy.Kind, err)
			}
			delivery.Retry.Retry(delivery.Step, rateLimitedAttempts, MaxRateLimitedAttempts, err.Error())
			delivery.Sleep(ctx, policy.Wait)
		default:
			return zero, WrapForFallback(policy.Kind, err)
		}
	}
}

// DraftRun は invalid-draft retry の loop を表す。Fetch は直前の失敗（無ければ空）を受けて原稿の断片を返す。
type DraftRun struct {
	Retry port.RetryReporter
	Build func(string) (models.ManuscriptDraft, error)
	Fetch func(context.Context, port.LastAttempt) (string, error)
}

func (d DraftRun) Run(ctx context.Context) (models.ManuscriptDraft, error) {
	var last port.LastAttempt
	for attempt := 1; attempt <= MaxDraftAttempts; attempt++ {
		raw, err := d.Fetch(ctx, last)
		if err != nil {
			return models.ManuscriptDraft{}, withLastAttempt(err, last)
		}
		draft, err := d.Build(raw)
		if err == nil {
			return draft, nil
		}
		last = port.LastAttempt{Raw: raw, BuildErr: err}
		if attempt < MaxDraftAttempts {
			d.Retry.Retry("write_manuscript_draft", attempt, MaxDraftAttempts, err.Error())
		}
	}
	return models.ManuscriptDraft{}, fmt.Errorf("%w: %w: %w", port.ErrDraftRejected, last.BuildErr, last)
}

// withLastAttempt は直前 attempt が invalid だった時だけ、その記録を err の chain へ足す。
func withLastAttempt(err error, last port.LastAttempt) error {
	if last.BuildErr == nil {
		return err
	}
	return fmt.Errorf("%w: %w", err, last)
}

// SleepContext は ctx が先に切れたらそちらを優先して待ちを中断する。
func SleepContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
