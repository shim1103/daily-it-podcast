//go:build r2smoke

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

// scope: Sociable Unit
// 実物境界: r2.ListObjectKeysSmoke / r2.GetObjectSmoke / r2.DeleteObjectSmoke が
//   signV4Get / buildObjectURL 等の既存 unexported ロジックへ委譲すること。
// double: 境界 I/O なしの seqRoundTripper（writer_sociable_unit_test.go と同型）。
// @invariant これら 3 関数は疎通確認専用であり、本番 EpisodeWriter / CompletedEpisodeLookup へ
//   メソッドとして追加しない（Interface Segregation を保つ）。

func TestListObjectKeysSmoke_returnsKeys_whenUpstreamSucceeds(t *testing.T) {
	t.Parallel()

	// Given: ListObjectsV2 が 2 key を返す
	rt := &seqRoundTripper{resps: []stubResp{
		{Status: http.StatusOK, Body: listObjectsV2XML("probe-1.json", "probe-1.mp3")},
	}}
	client := &http.Client{Transport: rt}

	// When: List する
	keys, err := r2.ListObjectKeysSmoke(context.Background(), client, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket)

	// Then: 2 key。GET method・Host・SigV4 Authorization
	if err != nil {
		t.Fatalf("ListObjectKeysSmoke: %v", err)
	}
	if len(keys) != 2 || keys[0] != "probe-1.json" || keys[1] != "probe-1.mp3" {
		t.Fatalf("keys = %v, want [probe-1.json probe-1.mp3]", keys)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(rt.calls))
	}
	if rt.calls[0].Method != http.MethodGet {
		t.Fatalf("method = %q, want GET", rt.calls[0].Method)
	}
	wantHost := testAccountID + ".r2.cloudflarestorage.com"
	if rt.calls[0].Host != wantHost {
		t.Fatalf("host = %q, want %q", rt.calls[0].Host, wantHost)
	}
	if !strings.HasPrefix(rt.calls[0].Auth, "AWS4-HMAC-SHA256 ") {
		t.Fatal("Authorization missing SigV4")
	}
}

func TestListObjectKeysSmoke_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	t.Parallel()

	// Given: List が常に 403 を返す（fail-fast 対象の 4xx）
	rt := &seqRoundTripper{resps: []stubResp{{Status: http.StatusForbidden, Body: "denied"}}}
	client := &http.Client{Transport: rt}

	// When: List する
	_, err := r2.ListObjectKeysSmoke(context.Background(), client, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket)

	// Then: *adaptererror.Error（r2: prefix）で secret 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertNoSecretLeak(t, err.Error())
}

func TestGetObjectSmoke_returnsBody_whenUpstreamSucceeds(t *testing.T) {
	t.Parallel()

	// Given: GetObject が 200 と body を返す
	rt := &seqRoundTripper{resps: []stubResp{{Status: http.StatusOK, Body: "probe-content"}}}
	client := &http.Client{Transport: rt}

	// When: Get する
	got, err := r2.GetObjectSmoke(context.Background(), client, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket, "probe-1.json")

	// Then: byte 一致。GET method・Host・path・SigV4
	if err != nil {
		t.Fatalf("GetObjectSmoke: %v", err)
	}
	if string(got) != "probe-content" {
		t.Fatalf("body = %q, want %q", got, "probe-content")
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(rt.calls))
	}
	wantPath := "/" + testBucket + "/probe-1.json"
	if rt.calls[0].Method != http.MethodGet || rt.calls[0].Path != wantPath {
		t.Fatalf("method/path = %q %q, want GET %q", rt.calls[0].Method, rt.calls[0].Path, wantPath)
	}
	if !strings.HasPrefix(rt.calls[0].Auth, "AWS4-HMAC-SHA256 ") {
		t.Fatal("Authorization missing SigV4")
	}
}

func TestGetObjectSmoke_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	t.Parallel()

	// Given: Get が常に 404 を返す（fail-fast 対象の 4xx）
	rt := &seqRoundTripper{resps: []stubResp{{Status: http.StatusNotFound, Body: "not found"}}}
	client := &http.Client{Transport: rt}

	// When: Get する
	_, err := r2.GetObjectSmoke(context.Background(), client, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket, "missing.json")

	// Then: *adaptererror.Error（r2: prefix）で secret・object key 非露出
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertNoSecretLeak(t, err.Error())
}

func TestDeleteObjectSmoke_succeeds_whenUpstreamReturns204(t *testing.T) {
	t.Parallel()

	// Given: DeleteObject が 204 を返す（S3 互換の正常応答）
	rt := &seqRoundTripper{resps: []stubResp{{Status: http.StatusNoContent}}}
	client := &http.Client{Transport: rt}

	// When: Delete する
	err := r2.DeleteObjectSmoke(context.Background(), client, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket, "probe-1.json")

	// Then: error なし。DELETE method・path・SigV4
	if err != nil {
		t.Fatalf("DeleteObjectSmoke: %v", err)
	}
	if len(rt.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(rt.calls))
	}
	wantPath := "/" + testBucket + "/probe-1.json"
	if rt.calls[0].Method != http.MethodDelete || rt.calls[0].Path != wantPath {
		t.Fatalf("method/path = %q %q, want DELETE %q", rt.calls[0].Method, rt.calls[0].Path, wantPath)
	}
	if !strings.HasPrefix(rt.calls[0].Auth, "AWS4-HMAC-SHA256 ") {
		t.Fatal("Authorization missing SigV4")
	}
}

func TestDeleteObjectSmoke_returnsInfrastructureError_whenUpstreamFails(t *testing.T) {
	t.Parallel()

	// Given: Delete が常に 500 を返す
	rt := &seqRoundTripper{resps: []stubResp{{Status: http.StatusInternalServerError, Body: "fail"}, {Status: http.StatusInternalServerError, Body: "fail"}}}
	client := &http.Client{Transport: rt}

	// When: Delete する
	err := r2.DeleteObjectSmoke(context.Background(), client, testAccessKeyID, testSecretAccessKey, testAccountID, testBucket, "probe-1.json")

	// Then: *adaptererror.Error（r2: prefix）で secret 非露出。5xx は有限 retry
	if err == nil {
		t.Fatal("expected error")
	}
	var infra *adaptererror.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *adaptererror.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "r2:") {
		t.Fatalf("Error() = %q, want prefix %q", infra.Error(), "r2:")
	}
	assertNoSecretLeak(t, err.Error())
	if len(rt.calls) != 2 {
		t.Fatalf("calls = %d, want 2 (retry once on 5xx)", len(rt.calls))
	}
}
