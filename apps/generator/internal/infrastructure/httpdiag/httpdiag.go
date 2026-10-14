// Package httpdiag は Driven Adapter が失敗診断へ添える HTTP 応答本文を
// 1 行の文字列へ落とす。secret 除去の責務は持たない（呼び出し側が
// 「どの header に credential を載せるから body 全出しで安全か」を保証する）。
package httpdiag

import "strings"

// BodySnippet は応答本文を 1 行の診断文字列へ落とす。
// 改行を空白へ潰し、長さ上限は設けない（quotaMetric 等の切り分け文が切れないようにする）。
func BodySnippet(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
