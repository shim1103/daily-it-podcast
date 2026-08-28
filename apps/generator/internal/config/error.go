package config

import "errors"

// runtime config validation の違反種別。errors.Is で分類する。
var (
	// ErrMissing はenvironment keyが未定義である。
	ErrMissing = errors.New("missing")
	// ErrEmpty はenvironment keyが定義済みだが空文字である。
	ErrEmpty = errors.New("empty")
	// ErrInvalidFormat は先頭または末尾にwhitespaceを含む等のformat違反である。
	ErrInvalidFormat = errors.New("invalid_format")
)
