package composition

import (
	"net/http"
	"os"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// httpTimeout は外向き HTTP Adapter が共有する Client の全体 timeout である。
const httpTimeout = 30 * time.Second

// sharedHTTPClient は全 HTTP Adapter が共有する *http.Client を返す。
//
// @ensure 戻りは適切な timeout を持つ標準 *http.Client。
func sharedHTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// sharedLookupEnv は全 command launcher / config.Load が共有する親環境アクセス手段を返す。
// production では os.LookupEnv。sharedHTTPClient と同じ production runtime 既定値。
//
// @ensure 戻りは os.LookupEnv。
func sharedLookupEnv() func(key string) (string, bool) {
	return os.LookupEnv
}

// sharedDisplayLocation は原稿 date を暦日化する表示タイムゾーンを解決して返す。
// production では constants.DisplayTimeZone を time.LoadLocation で解決する。sharedHTTPClient と同じ production runtime 既定値。
//
// @ensure 戻りは constants.DisplayTimeZone に対応する非 nil *time.Location。
// @invariant tzdata を読めない場合は panic する。表示タイムゾーンが引けない環境ではそもそも episode を出せないため起動前に落とす。
func sharedDisplayLocation() *time.Location {
	loc, err := time.LoadLocation(constants.DisplayTimeZone)
	if err != nil {
		panic("composition: load display location failed: " + err.Error())
	}
	return loc
}
