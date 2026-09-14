package runtime

import (
	"net/http"
	"time"
)

// httpTimeout は Drive / OAuth / 情報源 Adapter が共有する Client の全体 timeout である。
const httpTimeout = 30 * time.Second

// HTTPClient は短い request/response の HTTP Adapter が共有する *http.Client を返す。
//
// @ensure 戻りは httpTimeout を持つ標準 *http.Client。
func HTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// HTTPClientWithoutTimeout は Client.Timeout を置かない *http.Client を返す。
// request 全体の上限を Client ではなく ctx / process cancel、あるいは Adapter が自分で
// 付け直す timeout に委ねるときに使う。
//
// @ensure 戻りは全体 timeout を持たない標準 *http.Client。
func HTTPClientWithoutTimeout() *http.Client {
	return &http.Client{}
}
