// Package errors は Generator の Domain Error 実体と、層横断の Error 分類契約（Kinded / Kind*）を提供する。
package errors

// Kinded は Error 値が自分の External 分類（kind）と操作名（op）を名乗る契約である。
// why: delivery の classify が infra 具体型を instanceof 分岐せず、Error 側の宣言だけで
//
//	kind/op を採れるようにする（architecture/error-taxonomy §6）。
type Kinded interface {
	ErrorKind() string
	ErrorOp() string
}

// why: kind 語彙の typo は compile で捕まらないので定数へ固定する。delivery / infra / config が共有する。
const (
	KindDomain         = "domain"
	KindInfrastructure = "infrastructure"
	KindConfig         = "config"
)
