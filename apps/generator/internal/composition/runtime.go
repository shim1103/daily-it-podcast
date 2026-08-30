package composition

import (
	"net/http"
	"os"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// httpTimeout は GetX / Drive / OAuth など共有 HTTP Client の全体 timeout である。
const httpTimeout = 30 * time.Second

// geminiHTTPTimeout は Gemini TTS 1 呼び出しの Client 全体 timeout である。
// why: 共有 30s では長文朗読で awaiting headers が切れ、System が
// context deadline exceeded で落ちた（run 33310692613）。TTS は応答が重い。
const geminiHTTPTimeout = 120 * time.Second

// sharedHTTPClient は GetX / Drive / OAuth が共有する *http.Client を返す。
//
// @ensure 戻りは適切な timeout を持つ標準 *http.Client。
func sharedHTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// geminiHTTPClient は Gemini TTS 専用の *http.Client を返す。
//
// @ensure 戻りは geminiHTTPTimeout を持つ標準 *http.Client。
func geminiHTTPClient() *http.Client {
	return &http.Client{Timeout: geminiHTTPTimeout}
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
