package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

func TestFetchSourceItems_passesSinceAsNowMinusFetchWindow_whenNowGiven(t *testing.T) {
	// Given: 固定 now
	fake := &fakeItemSource{}
	uc := application.NewFetchSourceItems(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)
	wantSince := now.Add(-constants.FetchWindow)

	// When: Run を呼ぶ
	_, err := uc.Run(context.Background(), now)

	// Then: List は 1 回、since は now - FetchWindow
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("List calls = %d, want 1", len(fake.calls))
	}
	if !fake.calls[0].Equal(wantSince) {
		t.Fatalf("since = %v, want %v", fake.calls[0], wantSince)
	}
}

func TestFetchSourceItems_returnsItemsFromSource_whenListSucceeds(t *testing.T) {
	// Given: List が 2 件を返す
	occurred := time.Date(2024, 12, 10, 10, 0, 0, 0, time.UTC)
	want := []models.SourceItem{
		{SourceID: "x", OccurredAt: occurred, Context: "item_id: a1"},
		{SourceID: "x", OccurredAt: occurred.Add(time.Minute), Context: "item_id: a2"},
	}
	fake := &fakeItemSource{items: want}
	uc := application.NewFetchSourceItems(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: source の配列をそのまま返す
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].SourceID != want[i].SourceID || got[i].Context != want[i].Context || !got[i].OccurredAt.Equal(want[i].OccurredAt) {
			t.Fatalf("got[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFetchSourceItems_returnsEmptySlice_whenListReturnsEmpty(t *testing.T) {
	// Given: List が空 slice
	fake := &fakeItemSource{}
	uc := application.NewFetchSourceItems(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: 非 nil の空 slice
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got == nil {
		t.Fatal("got = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestFetchSourceItems_returnsErrorWithoutItems_whenListFails(t *testing.T) {
	// Given: List が失敗する
	boom := errors.New("list failed")
	fake := &fakeItemSource{
		items: []models.SourceItem{{SourceID: "x", Context: "item_id: a1"}},
		err:   boom,
	}
	uc := application.NewFetchSourceItems(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: その error。成功結果は返さない
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("List calls = %d, want 1", len(fake.calls))
	}
}
