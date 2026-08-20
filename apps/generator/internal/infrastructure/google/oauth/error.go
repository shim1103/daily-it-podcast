package oauth

// Error は Google OAuth refresh 呼び出しまたは応答変換の失敗。
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return "google oauth: " + e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func infraErr(op string, err error) error {
	return &Error{Op: op, Err: err}
}
