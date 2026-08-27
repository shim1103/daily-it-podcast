package config

// ViolationKind はruntime config validation違反の種類である。
type ViolationKind string

const (
	// ViolationMissing はmissingであり、environment keyが未定義の違反である。
	ViolationMissing ViolationKind = "missing"
	// ViolationEmpty はemptyであり、environment keyが定義済みだが空文字の違反である。
	ViolationEmpty ViolationKind = "empty"
	// ViolationInvalidFormat はwhitespace等のformat違反である。
	ViolationInvalidFormat ViolationKind = "invalid_format"
)

// Violation はkeyごとのruntime config validation違反である。
//
// @invariant raw valueを保持しない。
type Violation struct {
	Key  string
	Kind ViolationKind
}

// Error はGenerator runtime configのvalidation error契約である。
//
// @ensure Violationsはcallerが変更してもError内部へ影響しないdefensiveなlistを返す。
// @invariant Error文字列とViolationにraw valueを含めない。
type Error interface {
	error
	Violations() []Violation
}
