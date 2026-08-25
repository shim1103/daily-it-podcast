package agentsecrets

// Error は AgentSecrets secrettransport の失敗（Infrastructure Error）。
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return "agentsecrets: " + e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func infraErr(op string, err error) error {
	return &Error{Op: op, Err: err}
}
