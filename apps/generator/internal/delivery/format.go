// Package delivery は Generator CLI の Driving Adapter の出力面を持つ。
// error は Format が External 失敗行へ、非 error の観測（進捗・fallback 通知）は LogWriter が
// "generator: " 1 行 format へ写す。生 log をこの package の外へ出さない。
package delivery

import (
	"fmt"
	"strings"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
)

// why: 分類不能時のフォールバックだけ delivery が持つ。domain / infrastructure / config は
//
//	entities/errors の共有語彙（Kinded が名乗る）を参照する（architecture/error-taxonomy §5）。
const kindUnknown = "unknown"

// Format は Internal Error を CLI stderr 用の External 行へ写す。
//
// @require err は失敗経路の non-nil error。
// @ensure 戻りは kind 行を含む非空文字列。各行は "generator: " で始まる。
// @ensure secret 値を新たに挿入しない。message は err.Error() の1行化、cause は Unwrap 連鎖。
func Format(err error) string {
	if err == nil {
		return "generator: kind=" + kindUnknown + "\ngenerator: message=<nil>\n"
	}

	kind, op := classify(err)
	var b strings.Builder
	fmt.Fprintf(&b, "generator: kind=%s\n", kind)
	if op != "" {
		fmt.Fprintf(&b, "generator: op=%s\n", op)
	}
	fmt.Fprintf(&b, "generator: message=%s\n", oneLine(err.Error()))
	for i, cause := range causeLines(err) {
		fmt.Fprintf(&b, "generator: cause[%d]=%s\n", i, cause)
	}
	return b.String()
}

func classify(err error) (kind, op string) {
	// why: Unwrap 連鎖を1段ずつ外側から見て、最初に Kinded を名乗った Error の宣言をそのまま採る。
	for current := err; current != nil; current = unwrapOne(current) {
		if k, ok := current.(domainerrors.Kinded); ok {
			return k.ErrorKind(), k.ErrorOp()
		}
	}
	return kindUnknown, ""
}

func causeLines(err error) []string {
	var lines []string
	walkCauses(err, &lines)
	return lines
}

func walkCauses(err error, lines *[]string) {
	if err == nil {
		return
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		next := u.Unwrap()
		if next == nil {
			return
		}
		*lines = append(*lines, oneLine(next.Error()))
		walkCauses(next, lines)
		return
	}
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		for _, next := range u.Unwrap() {
			if next == nil {
				continue
			}
			*lines = append(*lines, oneLine(next.Error()))
			walkCauses(next, lines)
		}
	}
}

func unwrapOne(err error) error {
	if err == nil {
		return nil
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return u.Unwrap()
	}
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		errs := u.Unwrap()
		if len(errs) == 0 {
			return nil
		}
		return errs[0]
	}
	return nil
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
