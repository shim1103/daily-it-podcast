package constants_test

import (
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

func TestYesterdayHalfOpenWindow_returnsJSTYesterdayMidnightToTodayMidnight_whenNowIsCronMorning(t *testing.T) {
	t.Parallel()

	// Given: JST 05:00（cron 相当）。表示 TZ は Asia/Tokyo 固定
	jst := time.FixedZone("JST", 9*3600)
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, jst)

	// When: 昨日 half-open 窓を算出する
	since, until := constants.YesterdayHalfOpenWindow(now, jst)

	// Then: [昨日 00:00, 今日 00:00)
	wantSince := time.Date(2026, 9, 26, 0, 0, 0, 0, jst)
	wantUntil := time.Date(2026, 9, 27, 0, 0, 0, 0, jst)
	if !since.Equal(wantSince) {
		t.Fatalf("since = %v, want %v", since, wantSince)
	}
	if !until.Equal(wantUntil) {
		t.Fatalf("until = %v, want %v", until, wantUntil)
	}
}

func TestYesterdayHalfOpenWindow_usesDisplayCalendarDay_whenNowIsUTCCrossingJST(t *testing.T) {
	t.Parallel()

	// Given: UTC 2026-09-26 20:00 = JST 2026-09-27 05:00。窓は JST 暦日基準
	jst := time.FixedZone("JST", 9*3600)
	nowUTC := time.Date(2026, 9, 26, 20, 0, 0, 0, time.UTC)

	// When
	since, until := constants.YesterdayHalfOpenWindow(nowUTC, jst)

	// Then: JST 昨日〜今日の真夜中（instant 一致）
	wantSince := time.Date(2026, 9, 26, 0, 0, 0, 0, jst)
	wantUntil := time.Date(2026, 9, 27, 0, 0, 0, 0, jst)
	if !since.Equal(wantSince) {
		t.Fatalf("since = %v, want %v", since, wantSince)
	}
	if !until.Equal(wantUntil) {
		t.Fatalf("until = %v, want %v", until, wantUntil)
	}
}

func TestYesterdayHalfOpenWindow_staysOnPreviousCalendarDay_whenNowIsJustAfterMidnightJST(t *testing.T) {
	t.Parallel()

	// Given: JST 00:01。昨日はまだ「前々暦日 00:00〜前日 00:00」
	jst := time.FixedZone("JST", 9*3600)
	now := time.Date(2026, 9, 27, 0, 1, 0, 0, jst)

	// When
	since, until := constants.YesterdayHalfOpenWindow(now, jst)

	// Then
	wantSince := time.Date(2026, 9, 26, 0, 0, 0, 0, jst)
	wantUntil := time.Date(2026, 9, 27, 0, 0, 0, 0, jst)
	if !since.Equal(wantSince) {
		t.Fatalf("since = %v, want %v", since, wantSince)
	}
	if !until.Equal(wantUntil) {
		t.Fatalf("until = %v, want %v", until, wantUntil)
	}
}
