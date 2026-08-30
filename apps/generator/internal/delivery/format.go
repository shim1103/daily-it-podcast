// Package delivery は Generator CLI の Driving Adapter が使う External 失敗表現を組み立てる。
package delivery

import (
	"fmt"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
)

const (
	kindDomain         = "domain"
	kindConfig         = "config"
	kindInfrastructure = "infrastructure"
	kindUnknown        = "unknown"
)

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
	// why: errors.As を種類ごとに走らせると、内側の processenv が外側の cursorcli より先に当たる。
	for current := err; current != nil; current = unwrapOne(current) {
		switch e := current.(type) {
		case *domainerrors.Error:
			if e != nil {
				return kindDomain, e.Op
			}
		case *config.Error:
			if e != nil {
				return kindConfig, e.Key
			}
		case *config.Errors:
			if e != nil {
				return kindConfig, ""
			}
		case *cursorcli.Error:
			if e != nil {
				return kindInfrastructure, e.Op
			}
		case *processenv.Error:
			if e != nil {
				return kindInfrastructure, e.Op
			}
		case *gemini.Error:
			if e != nil {
				return kindInfrastructure, e.Op
			}
		case *getxapi.Error:
			if e != nil {
				return kindInfrastructure, e.Op
			}
		case *gdrive.Error:
			if e != nil {
				return kindInfrastructure, e.Op
			}
		case *oauth.Error:
			if e != nil {
				return kindInfrastructure, e.Op
			}
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
