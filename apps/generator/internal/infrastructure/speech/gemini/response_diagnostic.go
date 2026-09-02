package gemini

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// why: System 失敗の切り分けに応答本文が要るが、秘密は x-goog-api-key header にしか
//
//	載せない契約なので response body 全体を丸ごと出しても credential は漏れない。
//	長文 base64 で log を溢れさせないよう bounded にする。
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

// topLevelKeysHint は body を JSON object として読み、トップレベルキーを sort して
// " (top-level keys: [a, b, c])" 形式の追加情報へ落とす。
// body が JSON object でなければ空文字を返す（追加情報は付けない）。
func topLevelKeysHint(body []byte) string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return ""
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return fmt.Sprintf(" (top-level keys: [%s])", strings.Join(keys, ", "))
}
