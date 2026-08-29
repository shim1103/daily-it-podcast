package composition

import (
	"net/http"
	"os"
	"time"
)

// httpTimeout は外向き HTTP Adapter が共有する Client の全体 timeout である。
const httpTimeout = 30 * time.Second

// sharedHTTPClient は全 HTTP Adapter が共有する *http.Client を返す。
//
// @ensure 戻りは適切な timeout を持つ標準 *http.Client。
func sharedHTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// sharedLookupEnv は全 command launcher が共有する親環境アクセス手段を返す。
// production では os.LookupEnv。sharedHTTPClient と同じ production runtime 既定値。
//
// @ensure 戻りは os.LookupEnv。
func sharedLookupEnv() func(key string) (string, bool) {
	return os.LookupEnv
}
