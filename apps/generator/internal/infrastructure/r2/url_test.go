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
