package r2_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/r2"
)

func TestHasPair_returnsFalseNil_whenStubLookup(t *testing.T) {
	t.Parallel()

	// Given: flat ctor で組み立てた CompletedEpisodeLookup stub
	lookup := r2.NewCompletedEpisodeLookup(
		&http.Client{},
		testAccessKeyID,
		testSecretAccessKey,
		testAccountID,
		testBucket,
	)

	// When: HasPair する
	got, err := lookup.HasPair(context.Background(), "2026-08-31")

	// Then: false, nil（契約 freeze stub。List/Get 未実装）
	if err != nil {
		t.Fatalf("HasPair: %v", err)
	}
	if got {
		t.Fatal("HasPair = true, want false")
	}
}
