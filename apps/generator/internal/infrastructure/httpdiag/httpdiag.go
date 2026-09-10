// Package httpdiag は Driven Adapter が失敗診断へ添える HTTP 応答本文を
// 1 行・bounded の文字列へ落とす。secret 除去の責務は持たない（呼び出し側が
// 「どの header に credential を載せるから body 全出しで安全か」を保証する）。
package httpdiag

import "strings"

// why: 診断本文で log を溢れさせない上限。cursorapi / geminiapi / speech gemini が共有していた値。
const (
	maxRunes = 400
	ellipsis = "...(truncated)"
)

// BodySnippet は応答本文を 1 行・bounded の診断文字列へ落とす。
// 改行を空白へ潰し、maxRunes を超えたら "...(truncated)" を付ける。
//
// warn: ⚠️ 切り詰めは byte slice（s[:maxRunes]）で行う。maxRunes 超の本文が
//
//	マルチバイト境界で切れて不正 UTF-8 になりうるが、既存 3 実装と同挙動を維持する。
func BodySnippet(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if len(s) <= maxRunes {
		return s
	}
	return s[:maxRunes] + ellipsis
}
