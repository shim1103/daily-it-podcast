package runtime

import (
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// DisplayLocation は原稿 date を暦日化する表示タイムゾーンを解決して返す。
// production では constants.DisplayTimeZone を time.LoadLocation で解決する。
//
// @ensure 戻りは constants.DisplayTimeZone に対応する非 nil *time.Location。
// @invariant tzdata を読めない場合は panic する。表示タイムゾーンが引けない環境ではそもそも episode を出せないため起動前に落とす。
//
// why coverage: LoadLocation 失敗枝は定数 IANA 名が tzdata に無い環境でのみ到達する。
// testing-strategy/coverage §6 により、通常 CI / 開発機では到達不可能な分岐として専用 test を書かない。
func DisplayLocation() *time.Location {
	loc, err := time.LoadLocation(constants.DisplayTimeZone)
	if err != nil {
		panic("runtime: load display location failed: " + err.Error())
	}
	return loc
}
