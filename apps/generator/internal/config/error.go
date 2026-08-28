package config

import "strings"

// why: typo は compile で捕まらないので違反種別の語彙を定数へ固定する。
const (
	KindMissing       = "missing"
	KindEmpty         = "empty"
	KindInvalidFormat = "invalid_format"
)

// why: Domain "generator domain" / Infra package 名と同位置の層 prefix。3層の
// Error() を "<prefix>: <識別子>: <詳細>" で揃える。
const errorPrefix = "generator config"

type Error struct {
	Key  string
	Kind string
}

// Error は "<prefix>: <key>: <kind>" 形式を返す。
func (e *Error) Error() string {
	if e == nil {
		return errorPrefix + ": <nil>"
	}
	if e.Kind == "" {
		return errorPrefix + ": " + e.Key
	}
	return errorPrefix + ": " + e.Key + ": " + e.Kind
}

// why: application 層と対称に「層ごとに1つの生成 helper」。
func configErr(key, kind string) *Error {
	return &Error{Key: key, Kind: kind}
}

// Errors は複数の config 違反を Config の field 宣言順で束ねる。
type Errors struct {
	Violations []*Error
}

func (e *Errors) Error() string {
	if e == nil || len(e.Violations) == 0 {
		return errorPrefix + ": <nil>"
	}
	lines := make([]string, len(e.Violations))
	for i, v := range e.Violations {
		lines[i] = v.Error()
	}
	return strings.Join(lines, "\n")
}

// why: errors.As を個別 *Error へ到達させる（Issue §7-6）。
func (e *Errors) Unwrap() []error {
	if e == nil {
		return nil
	}
	out := make([]error, len(e.Violations))
	for i, v := range e.Violations {
		out[i] = v
	}
	return out
}
