package r2

import "testing"

func TestBuildObjectURL_returnsPathStyleURL_whenAccountIDAndBucketGiven(t *testing.T) {
	t.Parallel()

	// Given: accountID・bucket・objectName が揃っている
	// When: buildObjectURL を呼ぶ
	got, err := buildObjectURL("acc-1", "bucket-1", "ep-1.json")

	// Then: path-style の R2 Object URL が組み立つ
	if err != nil {
		t.Fatalf("buildObjectURL: %v", err)
	}
	want := "https://acc-1.r2.cloudflarestorage.com/bucket-1/ep-1.json"
	if got != want {
		t.Fatalf("got = %q, want %q", got, want)
	}
}

func TestBuildObjectURL_returnsError_whenAccountIDEmpty(t *testing.T) {
	t.Parallel()

	// Given: accountID が空
	// When: buildObjectURL を呼ぶ
	got, err := buildObjectURL("", "bucket-1", "ep-1.json")

	// Then: error になり URL は返らない
	if err == nil {
		t.Fatal("expected error")
	}
	if got != "" {
		t.Fatalf("got = %q, want empty", got)
	}
}

func TestBuildObjectURL_returnsError_whenBucketEmpty(t *testing.T) {
	t.Parallel()

	// Given: bucket が空
	// When: buildObjectURL を呼ぶ
	got, err := buildObjectURL("acc-1", "", "ep-1.json")

	// Then: error になり URL は返らない
	if err == nil {
		t.Fatal("expected error")
	}
	if got != "" {
		t.Fatalf("got = %q, want empty", got)
	}
}

func TestBuildObjectURL_escapesObjectName_whenContainsReservedCharacters(t *testing.T) {
	t.Parallel()

	// Given: objectName に URL 予約文字（space）を含む
	// When: buildObjectURL を呼ぶ
	got, err := buildObjectURL("acc-1", "bucket-1", "ep 1.json")

	// Then: path として正しく escape される
	if err != nil {
		t.Fatalf("buildObjectURL: %v", err)
	}
	want := "https://acc-1.r2.cloudflarestorage.com/bucket-1/ep%201.json"
	if got != want {
		t.Fatalf("got = %q, want %q", got, want)
	}
}

func TestBuildObjectURLFromBase_returnsPathUnderEndpoint_whenBaseAndBucketGiven(t *testing.T) {
	t.Parallel()

	// Given: local S3 の endpoint base（trailing slash 無し）
	// When: buildObjectURLFromBase を呼ぶ
	got, err := buildObjectURLFromBase("http://127.0.0.1:18787/cdn-cgi/local/r2/s3", "bucket-1", "ep-1.json")

	// Then: base 配下の path-style Object URL になる
	if err != nil {
		t.Fatalf("buildObjectURLFromBase: %v", err)
	}
	want := "http://127.0.0.1:18787/cdn-cgi/local/r2/s3/bucket-1/ep-1.json"
	if got != want {
		t.Fatalf("got = %q, want %q", got, want)
	}
}

func TestBuildObjectURLFromBase_trimsTrailingSlash_whenBaseHasTrailingSlash(t *testing.T) {
	t.Parallel()

	// Given: trailing slash 付き base
	// When: buildObjectURLFromBase を呼ぶ
	got, err := buildObjectURLFromBase("http://127.0.0.1:18787/cdn-cgi/local/r2/s3/", "bucket-1", "ep-1.json")

	// Then: 二重 slash にならない
	if err != nil {
		t.Fatalf("buildObjectURLFromBase: %v", err)
	}
	want := "http://127.0.0.1:18787/cdn-cgi/local/r2/s3/bucket-1/ep-1.json"
	if got != want {
		t.Fatalf("got = %q, want %q", got, want)
	}
}

func TestBuildObjectURLFromBase_returnsError_whenBaseEmpty(t *testing.T) {
	t.Parallel()

	// Given: base が空
	// When: buildObjectURLFromBase を呼ぶ
	got, err := buildObjectURLFromBase("", "bucket-1", "ep-1.json")

	// Then: error
	if err == nil {
		t.Fatal("expected error")
	}
	if got != "" {
		t.Fatalf("got = %q, want empty", got)
	}
}

func TestBuildListURLFromBase_appendsListTypeQuery_whenContinuationEmpty(t *testing.T) {
	t.Parallel()

	// Given: local S3 base と bucket
	// When: buildListURLFromBase を呼ぶ
	got, err := buildListURLFromBase("http://127.0.0.1:18787/cdn-cgi/local/r2/s3", "bucket-1", "")

	// Then: list-type=2 の ListObjectsV2 URL
	if err != nil {
		t.Fatalf("buildListURLFromBase: %v", err)
	}
	want := "http://127.0.0.1:18787/cdn-cgi/local/r2/s3/bucket-1?list-type=2"
	if got != want {
		t.Fatalf("got = %q, want %q", got, want)
	}
}
