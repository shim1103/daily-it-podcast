package r2

import "github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"

// retryLoop は有限回の attempt で fn を試行する共通 skeleton である。
// fn が retryable=false で終える、または attempt が尽きるまで直近の err を返す。
// why: list page 取得・get object・put object の retry skeleton が三重重複していたため共通化。
// retry は attempt 失敗時（次 attempt がある時のみ）に呼ぶ観測 hook。非 nil（呼び出し側の Constructor が保証する）。
func retryLoop[T any](retry port.RetryReporter, step string, attempts int, fn func() (retryable bool, val T, err error)) (T, error) {
	var last error
	var zero T
	for attempt := 1; attempt <= attempts; attempt++ {
		retryable, val, err := fn()
		if err == nil {
			return val, nil
		}
		last = err
		if !retryable {
			return zero, err
		}
		if attempt < attempts {
			retry.Retry(step, attempt, attempts, err.Error())
		}
	}
	return zero, last
}
