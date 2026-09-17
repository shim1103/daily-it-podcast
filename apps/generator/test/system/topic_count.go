package system

import (
	"os"
	"strconv"
	"strings"
)

const (
	// systemTestTopicCountEnv は system-test 実行時に topic 数を注入する環境変数名。
	systemTestTopicCountEnv = "SYSTEM_TEST_TOPIC_COUNT"
	// systemTestTopicCountDefault は疎通確認の費用と所要時間を抑える既定 topic 数。
	systemTestTopicCountDefault = 1
)

// systemTestTopicCount は環境変数から topic 数を読む。未設定または parse 失敗時は
// 疎通確認に必要な最小の 1 topic を使う。本番 topic 数は変更しない。
func systemTestTopicCount() int {
	raw := strings.TrimSpace(os.Getenv(systemTestTopicCountEnv))
	if raw == "" {
		return systemTestTopicCountDefault
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return systemTestTopicCountDefault
	}
	return n
}
