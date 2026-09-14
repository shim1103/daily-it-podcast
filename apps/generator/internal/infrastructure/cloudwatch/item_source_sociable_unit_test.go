package cloudwatch_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/cloudwatch"
)

func TestNewListItemSource_List_returnsNonNilEmptySlice(t *testing.T) {
	t.Parallel()

	source := cloudwatch.NewListItemSource(&http.Client{})
	got, err := source.List(context.Background(), time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("List() = nil, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(got))
	}
}
