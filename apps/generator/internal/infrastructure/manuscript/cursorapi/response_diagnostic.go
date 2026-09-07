package cursorapi

import "strings"

// why: System 失敗の切り分けに create 応答本文が要る。秘密は Authorization: Bearer header に
//
//	しか載せない契約なので、response body 全体を出しても credential は漏れない。
//	log を溢れさせないよう bounded にする（speech/gemini の bodySnippet と同型）。
const (
	bodySnippetMax      = 400
	bodySnippetEllipsis = "...(truncated)"
)

// bodySnippet は応答本文を 1 行・bounded の診断文字列へ落とす。
// 改行を空白へ潰し、bodySnippetMax rune を超えたら ellipsis を付ける。
func bodySnippet(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if len(s) <= bodySnippetMax {
		return s
	}
	return s[:bodySnippetMax] + bodySnippetEllipsis
}
