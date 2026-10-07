package writerretry

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// Scope: Sociable Unit
// 実物: writerretry の関数（ClassifyStatus・Run・DraftRun・WrapForFallback・SleepContext）
// Double: 待ちを観測する sleepSpy、Retry 通知を記録する retrySpy。HTTP は使わず、fetch は関数 Stub で与える。

type sleepSpy struct {
	waits []time.Duration
}

func (s *sleepSpy) sleep(_ context.Context, d time.Duration) {
	s.waits = append(s.waits, d)
}

type retryCall struct {
	step    string
	attempt int
	max     int
}

type retrySpy struct {
	calls []retryCall
}

func (s *retrySpy) Retry(step string, attempt, max int, _ string) {
	s.calls = append(s.calls, retryCall{step: step, attempt: attempt, max: max})
}

func newConfig() (Config, *sleepSpy, *retrySpy) {
	sleeps := &sleepSpy{}
	retries := &retrySpy{}
	return Config{Step: "fetch_step", Retry: retries, Sleep: sleeps.sleep}, sleeps, retries
}

var errFetch = errors.New("fetch failed")

func TestClassifyStatus_returnsPolicy_perStatusAndRetryAfter(t *testing.T) {
	t.Parallel()

	// Given: 失敗 status と Retry-After の組
	cases := []struct {
		name   string
		status int
		header http.Header
		want   Policy
	}{
		{"429 Retry-After 付き", http.StatusTooManyRequests, http.Header{"Retry-After": {"3"}}, Policy{Kind: RateLimited, Wait: 3 * time.Second}},
		{"429 Retry-After 過大はクランプ", http.StatusTooManyRequests, http.Header{"Retry-After": {"999999"}}, Policy{Kind: RateLimited, Wait: MaxRetryAfter}},
		{"429 Retry-After 無し", http.StatusTooManyRequests, nil, Policy{Kind: Fallback}},
		{"429 Retry-After 解釈不能", http.StatusTooManyRequests, http.Header{"Retry-After": {"soon"}}, Policy{Kind: Fallback}},
		{"429 Retry-After 0 は回復の明示なし", http.StatusTooManyRequests, http.Header{"Retry-After": {"0"}}, Policy{Kind: Fallback}},
		{"429 Retry-After 負数は回復の明示なし", http.StatusTooManyRequests, http.Header{"Retry-After": {"-1"}}, Policy{Kind: Fallback}},
		{"401 は credential 失効", http.StatusUnauthorized, nil, Policy{Kind: Fallback}},
		{"403 は credential 失効", http.StatusForbidden, nil, Policy{Kind: Fallback}},
		{"500", http.StatusInternalServerError, nil, Policy{Kind: Once}},
		{"503", http.StatusServiceUnavailable, nil, Policy{Kind: Once}},
		{"400 は bug", http.StatusBadRequest, nil, Policy{Kind: None}},
		{"404 は bug", http.StatusNotFound, nil, Policy{Kind: None}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// When: 分類する
			got := ClassifyStatus(tc.status, tc.header)

			// Then: 方針と待ちが決まる
			if got != tc.want {
				t.Fatalf("ClassifyStatus(%d, %v) = %+v, want %+v", tc.status, tc.header, got, tc.want)
			}
		})
	}
}

func TestWrapForFallback_wrapsSentinel_exceptForBug(t *testing.T) {
	t.Parallel()

	// Given: 失敗 error と各 kind
	for _, kind := range []Kind{Once, RateLimited, Fallback} {
		kind := kind
		t.Run("fallback へ渡す kind", func(t *testing.T) {
			t.Parallel()

			// When: wrap する
			got := WrapForFallback(kind, errFetch)

			// Then: 番兵と元 error の両方が辿れる
			if !errors.Is(got, port.ErrSourceExhausted) || !errors.Is(got, errFetch) {
				t.Fatalf("WrapForFallback(%v) = %v, want 番兵と元 error を含む", kind, got)
			}
		})
	}

	// When: bug の kind
	got := WrapForFallback(None, errFetch)

	// Then: wrap せずそのまま返す
	if got != errFetch {
		t.Fatalf("WrapForFallback(None) = %v, want 元 error そのまま", got)
	}
}

func TestRun_returnsValue_whenFetchSucceedsFirst(t *testing.T) {
	t.Parallel()

	// Given: 初回で成功する fetch
	cfg, sleeps, retries := newConfig()
	calls := 0

	// When: Run する
	got, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		return "ok", Policy{}, nil
	})

	// Then: 値を返し、待ちも通知もしない
	if err != nil || got != "ok" || calls != 1 || len(sleeps.waits) != 0 || len(retries.calls) != 0 {
		t.Fatalf("got=%q err=%v calls=%d sleeps=%v retries=%v", got, err, calls, sleeps.waits, retries.calls)
	}
}

func TestRun_retriesOnceThenSucceeds_whenPolicyIsOnce(t *testing.T) {
	t.Parallel()

	// Given: 1 回目だけ Once の失敗、2 回目は成功
	cfg, sleeps, _ := newConfig()
	calls := 0

	// When: Run する
	got, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		if calls == 1 {
			return "", Policy{Kind: Once}, errFetch
		}
		return "ok", Policy{}, nil
	})

	// Then: 即再試行して成功する（待たない）
	if err != nil || got != "ok" || calls != 2 || len(sleeps.waits) != 0 {
		t.Fatalf("got=%q err=%v calls=%d sleeps=%v", got, err, calls, sleeps.waits)
	}
}

func TestRun_wrapsFallbackAfterOneRetry_whenOnceKeepsFailing(t *testing.T) {
	t.Parallel()

	// Given: 常に Once の失敗
	cfg, _, _ := newConfig()
	calls := 0

	// When: Run する
	_, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		return "", Policy{Kind: Once}, errFetch
	})

	// Then: 再試行は 1 回だけ（calls == 2）で、fallback へ渡す番兵を返す
	if calls != 2 || !errors.Is(err, port.ErrSourceExhausted) || !errors.Is(err, errFetch) {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestRun_waitsAndNotifiesRetry_whenPolicyIsRateLimited(t *testing.T) {
	t.Parallel()

	// Given: Retry-After 3s 付きの失敗が 1 回、その後成功
	cfg, sleeps, retries := newConfig()
	calls := 0

	// When: Run する
	got, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		if calls == 1 {
			return "", Policy{Kind: RateLimited, Wait: 3 * time.Second}, errFetch
		}
		return "ok", Policy{}, nil
	})

	// Then: 指定の待ちで再試行し、step 名つきで Retry を通知する
	if err != nil || got != "ok" || len(sleeps.waits) != 1 || sleeps.waits[0] != 3*time.Second {
		t.Fatalf("got=%q err=%v sleeps=%v", got, err, sleeps.waits)
	}
	want := retryCall{step: "fetch_step", attempt: 1, max: MaxRateLimitedAttempts}
	if len(retries.calls) != 1 || retries.calls[0] != want {
		t.Fatalf("retries = %+v, want [%+v]", retries.calls, want)
	}
}

func TestRun_wrapsFallbackAfterMaxAttempts_whenRateLimitedKeepsFailing(t *testing.T) {
	t.Parallel()

	// Given: 常に Retry-After 付きの失敗
	cfg, sleeps, retries := newConfig()
	calls := 0

	// When: Run する
	_, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		return "", Policy{Kind: RateLimited, Wait: time.Second}, errFetch
	})

	// Then: MaxRateLimitedAttempts 回呼び、待ちと通知は 1 回少なく、fallback へ渡す番兵を返す
	if calls != MaxRateLimitedAttempts || len(sleeps.waits) != MaxRateLimitedAttempts-1 || len(retries.calls) != MaxRateLimitedAttempts-1 {
		t.Fatalf("calls=%d sleeps=%d retries=%d", calls, len(sleeps.waits), len(retries.calls))
	}
	if !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("err = %v, want 番兵を含む", err)
	}
}

func TestRun_wrapsFallbackWithoutRetry_whenPolicyIsFallback(t *testing.T) {
	t.Parallel()

	// Given: 即 fallback の失敗
	cfg, sleeps, _ := newConfig()
	calls := 0

	// When: Run する
	_, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		return "", Policy{Kind: Fallback}, errFetch
	})

	// Then: 再試行も待ちもせず、番兵を返す
	if calls != 1 || len(sleeps.waits) != 0 || !errors.Is(err, port.ErrSourceExhausted) {
		t.Fatalf("calls=%d sleeps=%v err=%v", calls, sleeps.waits, err)
	}
}

func TestRun_returnsPlainError_whenPolicyIsNone(t *testing.T) {
	t.Parallel()

	// Given: bug を示す失敗
	cfg, _, _ := newConfig()
	calls := 0

	// When: Run する
	_, err := Run(context.Background(), cfg, func(context.Context) (string, Policy, error) {
		calls++
		return "", Policy{}, errFetch
	})

	// Then: 再試行せず、番兵で wrap せずそのまま返す
	if calls != 1 || err != errFetch {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func okBuild(raw string) (models.ManuscriptDraft, error) {
	return models.ManuscriptDraft{Title: raw}, nil
}

func TestDraftRun_returnsDraft_whenFirstAttemptIsValid(t *testing.T) {
	t.Parallel()

	// Given: 初回で valid な draft になる
	retries := &retrySpy{}
	run := DraftRun{
		Retry: retries,
		Build: okBuild,
		Fetch: func(_ context.Context, last port.LastAttempt) (string, error) {
			if last.BuildErr != nil {
				t.Fatalf("初回の last = %+v, want 空", last)
			}
			return "raw", nil
		},
	}

	// When: 実行する
	got, err := run.Run(context.Background())

	// Then: draft を返し、通知しない
	if err != nil || got.Title != "raw" || len(retries.calls) != 0 {
		t.Fatalf("got=%+v err=%v retries=%v", got, err, retries.calls)
	}
}

func TestDraftRun_passesLastAttemptToNextFetch_whenBuildRejects(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は build が invalid、2 回目は valid
	retries := &retrySpy{}
	buildErr := errors.New("invalid draft")
	var lasts []port.LastAttempt
	attempt := 0
	run := DraftRun{
		Retry: retries,
		Build: func(raw string) (models.ManuscriptDraft, error) {
			if raw == "first" {
				return models.ManuscriptDraft{}, buildErr
			}
			return models.ManuscriptDraft{Title: raw}, nil
		},
		Fetch: func(_ context.Context, last port.LastAttempt) (string, error) {
			attempt++
			lasts = append(lasts, last)
			if attempt == 1 {
				return "first", nil
			}
			return "second", nil
		},
	}

	// When: 実行する
	got, err := run.Run(context.Background())

	// Then: 2 回目の fetch へ直前の raw と理由が渡り、step 名つきで通知される
	if err != nil || got.Title != "second" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if lasts[1].Raw != "first" || lasts[1].BuildErr != buildErr {
		t.Fatalf("2 回目の last = %+v", lasts[1])
	}
	want := retryCall{step: "write_manuscript_draft", attempt: 1, max: MaxDraftAttempts}
	if len(retries.calls) != 1 || retries.calls[0] != want {
		t.Fatalf("retries = %+v, want [%+v]", retries.calls, want)
	}
}

func TestDraftRun_wrapsDraftRejectedWithLastAttempt_whenEveryAttemptIsInvalid(t *testing.T) {
	t.Parallel()

	// Given: 毎回 invalid
	retries := &retrySpy{}
	buildErr := errors.New("always invalid")
	calls := 0
	run := DraftRun{
		Retry: retries,
		Build: func(string) (models.ManuscriptDraft, error) { return models.ManuscriptDraft{}, buildErr },
		Fetch: func(context.Context, port.LastAttempt) (string, error) {
			calls++
			return "raw", nil
		},
	}

	// When: 実行する
	_, err := run.Run(context.Background())

	// Then: MaxDraftAttempts 回で打ち切り、ErrDraftRejected と LastAttempt を chain に含む。最終 attempt は通知しない
	var last port.LastAttempt
	if calls != MaxDraftAttempts || !errors.Is(err, port.ErrDraftRejected) || !errors.Is(err, buildErr) || !errors.As(err, &last) || last.Raw != "raw" {
		t.Fatalf("calls=%d err=%v last=%+v", calls, err, last)
	}
	if len(retries.calls) != MaxDraftAttempts-1 {
		t.Fatalf("retries = %d, want %d", len(retries.calls), MaxDraftAttempts-1)
	}
}

func TestDraftRun_carriesLastAttempt_whenFetchFailsAfterRejection(t *testing.T) {
	t.Parallel()

	// Given: 1 回目は invalid、2 回目の fetch が失敗する
	fetchErr := errors.New("fetch broke")
	attempt := 0
	run := DraftRun{
		Retry: &retrySpy{},
		Build: func(string) (models.ManuscriptDraft, error) { return models.ManuscriptDraft{}, errors.New("invalid") },
		Fetch: func(context.Context, port.LastAttempt) (string, error) {
			attempt++
			if attempt == 1 {
				return "raw", nil
			}
			return "", fetchErr
		},
	}

	// When: 実行する
	_, err := run.Run(context.Background())

	// Then: 取得 error に直前の失敗が LastAttempt として chain に含まれる
	var last port.LastAttempt
	if !errors.Is(err, fetchErr) || !errors.As(err, &last) || last.Raw != "raw" {
		t.Fatalf("err=%v last=%+v", err, last)
	}
}

func TestDraftRun_returnsFetchErrorUnchanged_whenNoPriorRejection(t *testing.T) {
	t.Parallel()

	// Given: 初回の fetch が失敗する
	fetchErr := errors.New("fetch broke")
	run := DraftRun{
		Retry: &retrySpy{},
		Build: okBuild,
		Fetch: func(context.Context, port.LastAttempt) (string, error) { return "", fetchErr },
	}

	// When: 実行する
	_, err := run.Run(context.Background())

	// Then: 直前の失敗が無いので LastAttempt を足さず、そのまま返す
	var last port.LastAttempt
	if err != fetchErr || errors.As(err, &last) {
		t.Fatalf("err = %v, want fetchErr そのまま", err)
	}
}

func TestSleepContext_returnsEarly_whenContextIsDone(t *testing.T) {
	t.Parallel()

	// Given: 既に cancel 済みの ctx と長い待ち
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()

	// When: 待つ
	SleepContext(ctx, time.Hour)

	// Then: timer を待たずに即戻る
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("SleepContext blocked for %v", elapsed)
	}
}
