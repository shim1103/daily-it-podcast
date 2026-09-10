// Package adaptererror は Driven Adapter が返す Infrastructure Error を単一型で表す。
// 各 Adapter package は自前の Error 型を持たず、source を固定してこの型を生成する（DRY）。
// 分類契約（kind/op を名乗る interface）は entities/errors が所有する。
package adaptererror

import domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"

// Error は "<Source>: <Op>: <原因>" 形式の Infrastructure Error。
// Source は発生源（"gdrive" / "cursorapi" 等）、Op は操作、Err は原因。
type Error struct {
	Source string
	Op     string
	Err    error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil infra error>"
	}
	if e.Err == nil {
		return e.Source + ": " + e.Op
	}
	return e.Source + ": " + e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) ErrorKind() string {
	return domainerrors.KindInfrastructure
}

func (e *Error) ErrorOp() string {
	if e == nil {
		return ""
	}
	return e.Op
}

var _ domainerrors.Kinded = (*Error)(nil)

// New は Source を固定した Infrastructure Error を作る。
func New(source, op string, err error) *Error {
	return &Error{Source: source, Op: op, Err: err}
}
