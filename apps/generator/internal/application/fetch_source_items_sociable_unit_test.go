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

var testFetchLocation = time.FixedZone("JST", 9*3600)

func TestFetchSourceItems_passesYesterdayWindow_whenNowGiven(t *testing.T) {
	// Given: 固定 now（JST 05:00）と表示 Location
	fake := &fakeItemSource{}
	uc := application.NewFetchSourceItems(fake, testFetchLocation)
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, testFetchLocation)
	wantSince, wantUntil := constants.YesterdayHalfOpenWindow(now, testFetchLocation)

	// When: Run を呼ぶ
	_, err := uc.Run(context.Background(), now)

	// Then: List は 1 回、since/until は昨日 half-open 窓
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("List calls = %d, want 1", len(fake.calls))
	}
	if !fake.calls[0].since.Equal(wantSince) {
		t.Fatalf("since = %v, want %v", fake.calls[0].since, wantSince)
	}
	if !fake.calls[0].until.Equal(wantUntil) {
		t.Fatalf("until = %v, want %v", fake.calls[0].until, wantUntil)
	}
}

func TestFetchSourceItems_returnsItemsFromSource_whenListSucceeds(t *testing.T) {
	// Given: List が窓内 2 件を返す
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, testFetchLocation)
	since, _ := constants.YesterdayHalfOpenWindow(now, testFetchLocation)
	want := []models.SourceItem{
		{SourceID: "x", OccurredAt: since.Add(time.Hour), Summary: "item_id: a1"},
		{SourceID: "x", OccurredAt: since.Add(2 * time.Hour), Summary: "item_id: a2"},
	}
	fake := &fakeItemSource{items: want}
	uc := application.NewFetchSourceItems(fake, testFetchLocation)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: source の配列をそのまま返す（再 filter しない）
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].SourceID != want[i].SourceID || got[i].Summary != want[i].Summary || !got[i].OccurredAt.Equal(want[i].OccurredAt) {
			t.Fatalf("got[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFetchSourceItems_returnsEmptySlice_whenListReturnsEmpty(t *testing.T) {
	// Given: List が空 slice
	fake := &fakeItemSource{}
	uc := application.NewFetchSourceItems(fake, testFetchLocation)
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, testFetchLocation)

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
		items: []models.SourceItem{{SourceID: "x", Summary: "item_id: a1"}},
		err:   boom,
	}
	uc := application.NewFetchSourceItems(fake, testFetchLocation)
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, testFetchLocation)

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
