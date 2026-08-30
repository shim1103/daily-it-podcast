package gdrive

// Error は Google Drive 呼び出しまたは応答変換の失敗（Infrastructure Error）。
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return "gdrive: " + e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func infraErr(op string, err error) error {
	return &Error{Op: op, Err: err}
}
