package adaptererror

import (
	"errors"
	"testing"
)

func TestError_formatsSourceOpCause_whenCausePresent(t *testing.T) {
	t.Parallel()

	// Given: source / op / 原因つきの Infrastructure Error
	err := New("gdrive", "write", errors.New("quota"))

	// When: 文字列表現を採る
	got := err.Error()

	// Then: "<source>: <op>: <原因>" 形式
	if got != "gdrive: write: quota" {
		t.Fatalf("Error() = %q, want %q", got, "gdrive: write: quota")
	}
}

func TestError_omitsCauseSegment_whenCauseNil(t *testing.T) {
	t.Parallel()

	// Given: 原因 nil の Infrastructure Error
	err := New("cursorapi", "do", nil)

	// When: 文字列表現を採る
	got := err.Error()

	// Then: "<source>: <op>" のみ
	if got != "cursorapi: do" {
		t.Fatalf("Error() = %q, want %q", got, "cursorapi: do")
	}
}

func TestError_returnsNilPlaceholder_whenReceiverNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var err *Error

	// When: 文字列表現を採る
	got := err.Error()

	// Then: nil プレースホルダ
	if got != "<nil infra error>" {
		t.Fatalf("Error() = %q, want %q", got, "<nil infra error>")
	}
}

func TestUnwrap_returnsCause_whenCausePresent(t *testing.T) {
	t.Parallel()

	// Given: 原因つきの Infrastructure Error
	cause := errors.New("boom")
	err := New("geminiapi", "http_status", cause)

	// When: Unwrap する
	got := err.Unwrap()

	// Then: 原因 error がそのまま返る
	if got != cause {
		t.Fatalf("Unwrap() = %v, want %v", got, cause)
	}
}

func TestUnwrap_returnsNil_whenReceiverNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var err *Error

	// When: Unwrap する
	// Then: nil
	if got := err.Unwrap(); got != nil {
		t.Fatalf("Unwrap() = %v, want nil", got)
	}
}

func TestErrorKind_returnsInfrastructure(t *testing.T) {
	t.Parallel()

	// Given: 任意の Infrastructure Error
	err := New("lobsters", "list", errors.New("status 502"))

	// When / Then: kind は "infrastructure" 固定
	if got := err.ErrorKind(); got != "infrastructure" {
		t.Fatalf("ErrorKind() = %q, want %q", got, "infrastructure")
	}
}

func TestErrorOp_returnsOp_whenReceiverNonNil(t *testing.T) {
	t.Parallel()

	// Given: op を持つ Infrastructure Error
	err := New("itmedia", "fetch", errors.New("status 504"))

	// When / Then: ErrorOp は Op を返す
	if got := err.ErrorOp(); got != "fetch" {
		t.Fatalf("ErrorOp() = %q, want %q", got, "fetch")
	}
}

func TestErrorOp_returnsEmpty_whenReceiverNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var err *Error

	// When / Then: ErrorOp は空文字
	if got := err.ErrorOp(); got != "" {
		t.Fatalf("ErrorOp() = %q, want empty", got)
	}
}
