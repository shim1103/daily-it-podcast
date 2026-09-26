package r2

// retryLoop は有限回の attempt で fn を試行する共通 skeleton である。
// fn が retryable=false で終える、または attempt が尽きるまで直近の err を返す。
// why: list page 取得・get object・put object の retry skeleton が三重重複していたため共通化。
func retryLoop[T any](attempts int, fn func() (retryable bool, val T, err error)) (T, error) {
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
	}
	return zero, last
}
