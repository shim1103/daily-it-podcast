package processenv

import (
	"slices"
	"testing"
)

func TestBuildChildEnv_omitsUndefinedAllowlistName_whenLookupMisses(t *testing.T) {
	t.Parallel()

	// Given: allowlist の 1 名だけが lookup で見つかる
	lookup := func(key string) (string, bool) {
		if key == "PATH" {
			return "/processenv-test-bin", true
		}
		return "", false
	}

	// When: child env を組み立てる
	got := buildChildEnv(
		[]string{"PATH", "HOME"},
		"PROCESSENV_TEST_SECRET",
		"dummy-secret",
		lookup,
	)

	// Then: 未定義名は落ち、定義済みと secret だけが残る
	want := []string{
		"PATH=/processenv-test-bin",
		"PROCESSENV_TEST_SECRET=dummy-secret",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("buildChildEnv() = %v, want %v", got, want)
	}
}
