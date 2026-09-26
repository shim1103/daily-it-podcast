package r2

import (
	"errors"
	"testing"
)

// retryReporterSpy は port.RetryReporter を満たし、Retry 呼び出しを記録する Spy。
// retry が実際に発生する test（retryable な失敗が 2 回目以降の attempt へ進む場合）専用。
type retryReporterSpy struct {
	calls int
}

func (s *retryReporterSpy) Retry(step string, attempt, max int, reason string) {
	s.calls++
}

func TestRetryLoop_returnsValueWithoutRetry_whenFirstAttemptSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 1 回目で成功する fn
	calls := 0
	fn := func() (bool, int, error) {
		calls++
		return true, 42, nil
	}

	// When: retryLoop を実行する
	got, err := retryLoop[int](nil, "test_step", 3, fn)

	// Then: 1 回だけ呼ばれ、値がそのまま返る
	if err != nil {
		t.Fatalf("retryLoop: %v", err)
	}
	if got != 42 {
		t.Fatalf("got = %d, want 42", got)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRetryLoop_stopsImmediately_whenRetryableFalseBeforeAttemptsExhausted(t *testing.T) {
	t.Parallel()

	// Given: 1 回目で retryable=false のまま失敗する fn（attempts には余裕がある）
	calls := 0
	wantErr := errors.New("fail-fast")
	fn := func() (bool, int, error) {
		calls++
		return false, 0, wantErr
	}

	// When: attempts=5 で retryLoop を実行する
	got, err := retryLoop[int](nil, "test_step", 5, fn)

	// Then: retry せず 1 回で即終了し、zero 値と直近 error が返る
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if got != 0 {
		t.Fatalf("got = %d, want zero value 0", got)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (no retry when retryable=false)", calls)
	}
}

func TestRetryLoop_returnsLastError_whenAllAttemptsFail(t *testing.T) {
	t.Parallel()

	// Given: 常に retryable=true で失敗し続ける fn。attempt ごとに異なる error を返す
	calls := 0
	errs := []error{errors.New("err-1"), errors.New("err-2"), errors.New("err-3")}
	fn := func() (bool, int, error) {
		e := errs[calls]
		calls++
		return true, 0, e
	}

	// When: attempts=3 で retryLoop を実行する
	got, err := retryLoop[int](&retryReporterSpy{}, "test_step", 3, fn)

	// Then: attempts 分だけ呼ばれ、直近（最後）の error が返る
	if !errors.Is(err, errs[len(errs)-1]) {
		t.Fatalf("err = %v, want %v (last error)", err, errs[len(errs)-1])
	}
	if got != 0 {
		t.Fatalf("got = %d, want zero value 0", got)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (attempts exhausted)", calls)
	}
}

func TestRetryLoop_failsImmediately_whenAttemptsIsOne(t *testing.T) {
	t.Parallel()

	// Given: retryable=true を返す fn だが attempts=1 なので retry の余地がない
	calls := 0
	wantErr := errors.New("only-attempt")
	fn := func() (bool, int, error) {
		calls++
		return true, 0, wantErr
	}

	// When: attempts=1 で retryLoop を実行する
	got, err := retryLoop[int](nil, "test_step", 1, fn)

	// Then: 1 回だけ呼ばれ、その error がそのまま返る
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if got != 0 {
		t.Fatalf("got = %d, want zero value 0", got)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRetryLoop_succeedsOnLastAttempt_whenPriorAttemptsRetryableFail(t *testing.T) {
	t.Parallel()

	// Given: 2 回 retryable=true で失敗し、3 回目に成功する fn
	calls := 0
	fn := func() (bool, string, error) {
		calls++
		if calls < 3 {
			return true, "", errors.New("transient")
		}
		return true, "ok", nil
	}

	// When: attempts=3 で retryLoop を実行する
	got, err := retryLoop[string](&retryReporterSpy{}, "test_step", 3, fn)

	// Then: 3 回目で成功し、値が返る
	if err != nil {
		t.Fatalf("retryLoop: %v", err)
	}
	if got != "ok" {
		t.Fatalf("got = %q, want %q", got, "ok")
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}
