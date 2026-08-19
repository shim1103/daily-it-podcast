package errors

// InvalidManuscript は原稿が manuscript.schema.json に適合しない。
type InvalidManuscript struct {
	Err error
}

func (e *InvalidManuscript) Error() string {
	return "invalid manuscript: " + e.Err.Error()
}

func (e *InvalidManuscript) Unwrap() error {
	return e.Err
}
