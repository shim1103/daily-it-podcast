package r2_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

func newStubLookup(rt *seqRoundTripper) *r2.CompletedEpisodeLookup {
	// why: 5xx / 429 / network error から再試行する test と成功一発の test を共有する。
	//      呼ばれるかどうかをテストごとに見極めず、常に Spy を渡して安全に倒す。
	return r2.NewCompletedEpisodeLookup(
		&http.Client{Transport: rt},
		testAccessKeyID,
		testSecretAccessKey,
		testAccountID,
		testBucket,
		&retryReporterSpy{},
	)
}

func listObjectsV2XML(keys ...string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`)
	for _, k := range keys {
		b.WriteString("<Contents><Key>")
		b.WriteString(k)
		b.WriteString("</Key></Contents>")
	}
	b.WriteString(`<IsTruncated>false</IsTruncated>`)
	b.WriteString(`</ListBucketResult>`)
	return b.String()
}

func TestHasPair_returnsTrue_whenSameStemJsonAndMp3MatchDate(t *testing.T) {
	t.Parallel()

	// Given: 同 stem の json+mp3。json の date が照会日と一致
	const stem = "ep-complete-1"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem+".json", stem+".mp3")},
		{Status: http.StatusOK, Body: `{"date":"2026-08-31","episodeId":"` + stem + `"}`},
	}}
	lookup := newStubLookup(rt)

	// When: 同日で照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: true。List と Get が走る
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if !got {
		t.Fatal("HasPair = false, want true")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(rt.calls))
	}
	if rt.calls[0].Method != http.MethodGet {
		t.Fatalf("list method = %q, want GET", rt.calls[0].Method)
	}
	if !strings.Contains(rt.calls[0].Path, "/"+testBucket) {
		t.Fatalf("list path = %q", rt.calls[0].Path)
	}
	if rt.calls[1].Method != http.MethodGet || !strings.HasSuffix(rt.calls[1].Path, "/"+stem+".json") {
		t.Fatalf("get = %q %q", rt.calls[1].Method, rt.calls[1].Path)
	}
}

func TestHasPair_returnsFalse_whenJsonOnly(t *testing.T) {
	t.Parallel()

	// Given: 同日 date になり得る json のみ（対応 mp3 無し）
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML("ep-json-only.json")},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false。Get しない（完成候補が無い）
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1 (list only)", len(rt.calls))
	}
}

func TestHasPair_returnsFalse_whenMp3Only(t *testing.T) {
	t.Parallel()

	// Given: mp3 のみ
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML("ep-mp3-only.mp3")},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}

func TestHasPair_returnsFalse_whenNoObjects(t *testing.T) {
	t.Parallel()

	// Given: bucket が空
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML()},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}

func TestHasPair_returnsFalse_whenPairDateDiffers(t *testing.T) {
	t.Parallel()

	// Given: 完成ペアはあるが date が別日
	const stem = "ep-other-day"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem+".json", stem+".mp3")},
		{Status: http.StatusOK, Body: `{"date":"2026-08-30","episodeId":"` + stem + `"}`},
	}}
	lookup := newStubLookup(rt)

	// When: 別日で照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}

func TestHasPair_retriesOnce_whenListReturns5xxThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: List が 500→200（空）。Get は不要
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusInternalServerError, Body: "fail"},
		{Status: http.StatusOK, Body: listObjectsV2XML()},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: 有限 retry 後に false, nil。List は 2 回
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(rt.calls))
	}
}

func TestHasPair_retriesOnce_whenListReturns429ThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: List が 429→200（空）
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusTooManyRequests},
		{Status: http.StatusOK, Body: listObjectsV2XML()},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: 有限 retry 後に成功
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(rt.calls))
	}
}

func TestHasPair_retriesOnce_whenNetworkFailsThenSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 初回 network fail、続く List は空 200
	rt := &seqRoundTripper{resps: []stubResp{
		{Err: errors.New("connection reset")},
		{Status: http.StatusOK, Body: listObjectsV2XML()},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: network 有限 retry 後に false, nil
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(rt.calls))
	}
}

func TestHasPair_failsFastWithoutRetry_whenListReturns4xxOtherThan429(t *testing.T) {
	t.Parallel()

	// Given: List が 403
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusForbidden, Body: "denied"},
		{Status: http.StatusOK, Body: listObjectsV2XML()}, // 到達しない想定
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: fail-fast で 1 回だけ。*adaptererror.Error・secret/key 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q", infra.Error())
	}
	assertNoSecretLeak(t, err.Error())
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1 (fail-fast)", len(rt.calls))
	}
}

func TestHasPair_failsFastWithoutRetry_whenListReturns201(t *testing.T) {
	t.Parallel()

	// Given: List（GetObjectsV2/GET）が 201 を返す。ListObjectsV2 は正常時常に 200 固定であり、
	// 201 は PUT/POST 系専用のため List 応答としては異常値として扱う想定
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusCreated, Body: listObjectsV2XML()},
		{Status: http.StatusOK, Body: listObjectsV2XML()}, // 到達しない想定（200 固定判定は retry 対象にしない）
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: fail-fast で 1 回だけ。*adaptererror.Error・secret/key 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1 (fail-fast, 200 固定判定なので 201 は success 扱いしない)", len(rt.calls))
	}
}

func TestHasPair_failsFastWithoutRetry_whenGetReturns201(t *testing.T) {
	t.Parallel()

	// Given: List は完成ペア候補を返すが、Get（GetObject/GET）が 201 を返す。
	// GetObject は正常時常に 200 固定であり、201 は Get 応答としては異常値として扱う想定
	const stem = "ep-get-201"
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML(stem+".json", stem+".mp3")},
		{Status: http.StatusCreated, Body: `{"date":"2026-08-31"}`},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: fail-fast で Get 1 回だけ（List 1 + Get 1 = 2）。*adaptererror.Error・secret/key 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2 (list + fail-fast get, 200 固定判定なので 201 は success 扱いしない)", len(rt.calls))
	}
}

func TestHasPair_returnsInfrastructureErrorWithoutSecrets_when5xxExhaustsRetries(t *testing.T) {
	t.Parallel()

	// Given: List が常に 502
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusBadGateway},
		{Status: http.StatusBadGateway},
	}}
	lookup := newStubLookup(rt)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: *adaptererror.Error・secret 非露出。List は 2 回で打ち切る
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2 (finite retry)", len(rt.calls))
	}
}

func TestHasPair_returnsInfrastructureError_whenClientNil(t *testing.T) {
	t.Parallel()

	// Given: http.Client が nil
	lookup := r2.NewCompletedEpisodeLookup(nil, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket, nil)

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: *adaptererror.Error かつ secret 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	assertNoSecretLeak(t, err.Error())
}

func TestHasPair_returnsInfrastructureError_whenLookupNil(t *testing.T) {
	t.Parallel()

	// Given: nil の *CompletedEpisodeLookup
	var lookup *r2.CompletedEpisodeLookup

	// When: 照会する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: *adaptererror.Error（r2: prefix）
	if err == nil {
		t.Fatal("expected error")
	}
	if got {
		t.Fatal("HasPair = true on error, want false")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q", infra.Error())
	}
}
