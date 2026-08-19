package twitterapiio_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/twitterapiio"
)

func TestList_returnsEmptySlice_whenClientPresent(t *testing.T) {
	// Given: client がある Stub
	source := twitterapiio.NewPostSource(&agentsecrets.Client{})
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: 空 slice（vendor 取得は todo）
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got == nil {
		t.Fatal("got = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestList_returnsInfrastructureError_whenClientNil(t *testing.T) {
	// Given: client が nil
	source := twitterapiio.NewPostSource(nil)
	since := time.Date(2024, 12, 10, 0, 0, 0, 0, time.UTC)

	// When: List を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: Infrastructure Error。Error / Unwrap が観測できる
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	var infra *twitterapiio.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *twitterapiio.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "twitterapiio:") {
		t.Fatalf("Error() = %q, want prefix twitterapiio:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}
