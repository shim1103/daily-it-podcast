package constants

import "time"

// YesterdayHalfOpenWindow は表示 Location の「昨日」暦日を half-open 区間 [since, until) で返す。
// since = 昨日 00:00:00、until = 今日 00:00:00（いずれも loc）。長さは常に 24h。
//
// @require loc != nil。
// @ensure since / until は loc の wall clock で真夜中。until.Sub(since) == 24h。
// @ensure now の時刻・実行遅延に依存せず、now を loc で見た暦日の「昨日」だけを返す。
// @invariant time.Now を呼ばない（純関数）。OccurredAt 自体は触らない。
func YesterdayHalfOpenWindow(now time.Time, loc *time.Location) (since, until time.Time) {
	local := now.In(loc)
	todayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	return yesterdayStart, todayStart
}

// OccurredInHalfOpen は t ∈ [since, until) なら true。
// ItemSource Adapter が OccurredAt を窓へ落とすときに使う。
//
// @ensure !t.Before(since) && t.Before(until)。
func OccurredInHalfOpen(t, since, until time.Time) bool {
	return !t.Before(since) && t.Before(until)
}
