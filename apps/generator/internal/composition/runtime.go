package composition

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// httpTimeout は Drive / OAuth / 情報源 Adapter が共有する Client の全体 timeout である。
const httpTimeout = 30 * time.Second

// sharedHTTPClient は短い request/response の HTTP Adapter が共有する *http.Client を返す。
//
// @ensure 戻りは httpTimeout を持つ標準 *http.Client。
func sharedHTTPClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

// sharedHTTPClientWithoutTimeout は Client.Timeout を置かない *http.Client を返す。
// request 全体の上限を Client ではなく ctx / process cancel、あるいは Adapter が自分で
// 付け直す timeout に委ねるときに使う（Cursor Cloud Agents の長い待ち、Gemini TTS の長文朗読など）。
// 短い request/response には sharedHTTPClient を使う。
//
// @ensure 戻りは全体 timeout を持たない標準 *http.Client。
func sharedHTTPClientWithoutTimeout() *http.Client {
	return &http.Client{}
}

// logManuscriptSourceSwitched は原稿の取得元が primary から secondary へ切り替わったことを
// 1 行だけ stderr へ出す。manuscript.NewTextWriter の onFallback として渡す。
// why: generator は構造化 log 基盤を持たないため、一時的に標準 log で stderr へ WARN 相当 1 行を出す。
//
//	Cursor の利用枠喪失で毎日 Gemini に落ちている状態を、後から GHA run log で気づけるようにする。
//	secret・brief 本文・token は出さない（DEPLOY.md の log 方針）。恒久化（構造化 log・能動通知）は別判断。
func logManuscriptSourceSwitched() {
	log.Print("manuscript text writer switched: from=cursor to=gemini reason=source_exhausted")
}

// sharedLookupEnv は config.Load が使う親環境アクセス手段を返す。
// production では os.LookupEnv。
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
