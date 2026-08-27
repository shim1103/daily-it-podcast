package config

import "fmt"

// Secret は表示時に値を露出しないopaqueなcredential契約である。
//
// @ensure StringとGoStringはraw valueを返さず、Revealだけがraw valueを返す。
// @invariant Secretの具体的な保存方法を利用側へ公開しない。
type Secret interface {
	fmt.Stringer
	fmt.GoStringer
	Reveal() string
}
