package gemini

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

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
