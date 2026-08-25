package errors

import "fmt"

// InvalidManuscriptDraft は TextWriter の text 断片を ManuscriptDraft へ解釈できない。
type InvalidManuscriptDraft struct {
	Err error
}

func (e *InvalidManuscriptDraft) Error() string {
	if e == nil || e.Err == nil {
		return "manuscript draft is invalid"
	}
	return fmt.Sprintf("manuscript draft is invalid: %v", e.Err)
}

func (e *InvalidManuscriptDraft) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
