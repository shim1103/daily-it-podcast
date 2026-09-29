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

func TestFetchSourceItems_passesYesterdaySince_whenNowGiven(t *testing.T) {
	// Given: 固定 now（JST 05:00）と表示 Location
	fake := &fakeItemSource{}
	uc := application.NewFetchSourceItems(fake, testFetchLocation)
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, testFetchLocation)
	wantSince, _ := constants.YesterdayHalfOpenWindow(now, testFetchLocation)

	// When: Run を呼ぶ
	_, err := uc.Run(context.Background(), now)

	// Then: List は 1 回、since は昨日 00:00（JST）
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

func TestFetchSourceItems_excludesItemsAtOrAfterUntil_whenListReturnsBorderItems(t *testing.T) {
	// Given: since ちょうど・until ちょうど・until 以降の item を混ぜた List 結果
	now := time.Date(2026, 9, 27, 5, 0, 0, 0, testFetchLocation)
	since, until := constants.YesterdayHalfOpenWindow(now, testFetchLocation)
	inWindow := models.SourceItem{SourceID: "x", OccurredAt: since.Add(time.Hour), Summary: "in"}
	atSince := models.SourceItem{SourceID: "x", OccurredAt: since, Summary: "at-since"}
	atUntil := models.SourceItem{SourceID: "x", OccurredAt: until, Summary: "at-until"}
	afterUntil := models.SourceItem{SourceID: "x", OccurredAt: until.Add(time.Hour), Summary: "after-until"}
	fake := &fakeItemSource{items: []models.SourceItem{inWindow, atSince, atUntil, afterUntil}}
	uc := application.NewFetchSourceItems(fake, testFetchLocation)

	// When
	got, err := uc.Run(context.Background(), now)

	// Then: [since, until) のみ。atUntil / afterUntil は落ちる
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Summary != "in" || got[1].Summary != "at-since" {
		t.Fatalf("got = %+v, want in then at-since", got)
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

	// Then: source の配列をそのまま返す
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
