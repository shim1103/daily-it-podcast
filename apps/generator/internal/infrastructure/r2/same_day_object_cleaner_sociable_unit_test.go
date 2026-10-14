package r2_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

func newStubCleaner(rt *seqRoundTripper) *r2.SameDayObjectCleaner {
	return r2.NewSameDayObjectCleaner(
		&http.Client{Transport: rt},
		testAccessKeyID,
		testSecretAccessKey,
		testAccountID,
		testBucket,
		&retryReporterSpy{},
	)
}

func objectPath(objectName string) string {
	return "/" + testBucket + "/" + objectName
}

func deletePaths(calls []stubCall) []string {
	var deletes []string
	for _, c := range calls {
		if c.Method == http.MethodDelete {
			deletes = append(deletes, c.Path)
		}
	}
	return deletes
}

func assertDeletePathSet(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("DELETE paths = %q, want set %q", got, want)
	}
	counts := make(map[string]int, len(want))
	for _, p := range want {
		counts[p]++
	}
	for _, p := range got {
		counts[p]--
		if counts[p] < 0 {
			t.Fatalf("DELETE paths = %q, want set %q", got, want)
		}
	}
}

func TestRemoveForDate_deletesBothJsonAndMp3_whenSameDayCompletePair(t *testing.T) {
	t.Parallel()

	// Given: 同日の完成ペア（json+mp3）。date 一致
	const stem = "ep-complete"
	const date = "2026-10-14"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem+".json", stem+".mp3")},
		{Status: http.StatusOK, Body: `{"date":"` + date + `","episodeId":"` + stem + `"}`},
		{Status: http.StatusNoContent},
		{Status: http.StatusNoContent},
	}}
	cleaner := newStubCleaner(rt)

	// When: 同日で RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), date)

	// Then: removed=1。json と mp3 の DELETE path が両方走る
	if err != nil {
		t.Fatalf("RemoveForDate: %v", err)
	}
	if n != 1 {
		t.Fatalf("removed = %d, want 1", n)
	}
	assertDeletePathSet(t, deletePaths(rt.calls), objectPath(stem+".json"), objectPath(stem+".mp3"))
}

func TestRemoveForDate_deletesJsonOnlySameDayRemnant_whenSiblingMp3Returns404(t *testing.T) {
	t.Parallel()

	// Given: 同日 json のみ（不完全残骸）。sibling mp3 は DELETE で 404
	const stem = "ep-json-only"
	const date = "2026-10-14"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem + ".json")},
		{Status: http.StatusOK, Body: `{"date":"` + date + `","episodeId":"` + stem + `"}`},
		{Status: http.StatusNoContent},
		{Status: http.StatusNotFound},
	}}
	cleaner := newStubCleaner(rt)

	// When: 同日で RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), date)

	// Then: removed=1。json DELETE + sibling mp3 DELETE（404 は成功）
	if err != nil {
		t.Fatalf("RemoveForDate: %v", err)
	}
	if n != 1 {
		t.Fatalf("removed = %d, want 1", n)
	}
	assertDeletePathSet(t, deletePaths(rt.calls), objectPath(stem+".json"), objectPath(stem+".mp3"))
}

func TestRemoveForDate_leavesOtherDayUntouched_whenSameDayAndOtherDayPairsExist(t *testing.T) {
	t.Parallel()

	// Given: 同日ペアと別日ペアが共存
	const stemMatch = "ep-same-day"
	const stemOther = "ep-other-day"
	const date = "2026-10-14"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(
			stemMatch+".json", stemMatch+".mp3",
			stemOther+".json", stemOther+".mp3",
		)},
		{Status: http.StatusOK, Body: `{"date":"` + date + `","episodeId":"` + stemMatch + `"}`},
		{Status: http.StatusNoContent},
		{Status: http.StatusNoContent},
		{Status: http.StatusOK, Body: `{"date":"2026-10-13","episodeId":"` + stemOther + `"}`},
	}}
	cleaner := newStubCleaner(rt)

	// When: 同日で RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), date)

	// Then: removed=1。DELETE は同日 stem のみ（別日 path は無い）
	if err != nil {
		t.Fatalf("RemoveForDate: %v", err)
	}
	if n != 1 {
		t.Fatalf("removed = %d, want 1", n)
	}
	assertDeletePathSet(t, deletePaths(rt.calls), objectPath(stemMatch+".json"), objectPath(stemMatch+".mp3"))
}

func TestRemoveForDate_leavesMp3Only_whenJsonAbsent(t *testing.T) {
	t.Parallel()

	// Given: mp3 のみ（json が無いので date が分からない）
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML("ep-orphan.mp3")},
	}}
	cleaner := newStubCleaner(rt)

	// When: RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), "2026-10-14")

	// Then: removed=0。DELETE は 0 回
	if err != nil {
		t.Fatalf("RemoveForDate: %v", err)
	}
	if n != 0 {
		t.Fatalf("removed = %d, want 0", n)
	}
	assertDeletePathSet(t, deletePaths(rt.calls))
}

func TestRemoveForDate_returnsZero_whenNoMatchingDate(t *testing.T) {
	t.Parallel()

	// Given: list 上の json は別日だけ
	const stem = "ep-only-other"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem+".json", stem+".mp3")},
		{Status: http.StatusOK, Body: `{"date":"2026-10-13","episodeId":"` + stem + `"}`},
	}}
	cleaner := newStubCleaner(rt)

	// When: 照会日に一致しない date で RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), "2026-10-14")

	// Then: removed=0。DELETE は 0 回
	if err != nil {
		t.Fatalf("RemoveForDate: %v", err)
	}
	if n != 0 {
		t.Fatalf("removed = %d, want 0", n)
	}
	assertDeletePathSet(t, deletePaths(rt.calls))
}

func TestRemoveForDate_failsFastWithoutRetry_whenDeleteReturns4xxOtherThan404(t *testing.T) {
	t.Parallel()

	// Given: 同日 json。json DELETE が 403（404 以外の 4xx。retry 対象外）
	const stem = "ep-del-4xx"
	const date = "2026-10-14"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem + ".json")},
		{Status: http.StatusOK, Body: `{"date":"` + date + `","episodeId":"` + stem + `"}`},
		{Status: http.StatusForbidden, Body: "denied"},
		{Status: http.StatusNoContent}, // 到達しない想定
	}}
	cleaner := newStubCleaner(rt)

	// When: RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), date)

	// Then: *adaptererror.Error。DELETE は 1 回で fail-fast。removed は 0
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if n != 0 {
		t.Fatalf("removed = %d, want 0", n)
	}
	assertDeletePathSet(t, deletePaths(rt.calls), objectPath(stem+".json"))
}

func TestRemoveForDate_retriesOnce_whenDeleteReturns5xxThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 同日完成ペア。json DELETE が 500→204、mp3 DELETE は 204
	const stem = "ep-del-5xx"
	const date = "2026-10-14"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem+".json", stem+".mp3")},
		{Status: http.StatusOK, Body: `{"date":"` + date + `","episodeId":"` + stem + `"}`},
		{Status: http.StatusInternalServerError, Body: "fail"},
		{Status: http.StatusNoContent},
		{Status: http.StatusNoContent},
	}}
	cleaner := newStubCleaner(rt)

	// When: RemoveForDate する
	n, err := cleaner.RemoveForDate(context.Background(), date)

	// Then: 有限 retry 後に成功。json DELETE path が 2 回、mp3 が 1 回
	if err != nil {
		t.Fatalf("RemoveForDate: %v", err)
	}
	if n != 1 {
		t.Fatalf("removed = %d, want 1", n)
	}
	deletes := deletePaths(rt.calls)
	if len(deletes) != 3 {
		t.Fatalf("DELETE count = %d paths=%q, want 3", len(deletes), deletes)
	}
	if deletes[0] != objectPath(stem+".json") || deletes[1] != objectPath(stem+".json") {
		t.Fatalf("json retry paths = %q, %q", deletes[0], deletes[1])
	}
	if deletes[2] != objectPath(stem+".mp3") {
		t.Fatalf("mp3 path = %q", deletes[2])
	}
}

func TestRemoveForDate_returnsInfrastructureErrorWithoutSecrets_whenDeleteFails(t *testing.T) {
	t.Parallel()

	// Given: 同日 json。json DELETE が 403 で失敗する
	const stem = "ep-del-secret"
	const date = "2026-10-14"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem + ".json")},
		{Status: http.StatusOK, Body: `{"date":"` + date + `","episodeId":"` + stem + `"}`},
		{Status: http.StatusForbidden, Body: "denied"},
	}}
	cleaner := newStubCleaner(rt)

	// When: RemoveForDate する
	_, err := cleaner.RemoveForDate(context.Background(), date)

	// Then: error。Error() に secret / resource 実値を載せない
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
}
