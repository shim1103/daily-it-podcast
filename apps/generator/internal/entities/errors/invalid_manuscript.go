package errors

import "fmt"

// InvalidManuscript は manuscript schema に適合しない原稿である。
type InvalidManuscript struct {
	Err error
}

func (e *InvalidManuscript) Error() string {
	if e == nil || e.Err == nil {
		return "manuscript is invalid"
	}
	return fmt.Sprintf("manuscript is invalid: %v", e.Err)
}

func (e *InvalidManuscript) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
