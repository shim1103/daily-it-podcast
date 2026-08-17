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

func withWatchUserIDs(t *testing.T, ids []string) {
	t.Helper()
	original := append([]string(nil), constants.WatchUserIDs...)
	constants.WatchUserIDs = ids
	t.Cleanup(func() { constants.WatchUserIDs = original })
}

func TestFetchWatchedPosts_callsListByUserForEachID_whenMultipleWatchUserIDs(t *testing.T) {
	// Given: WatchUserIDs が 2 名、Fake は空結果
	withWatchUserIDs(t, []string{"user-a", "user-b"})
	fake := &fakePostSource{posts: map[string][]models.Post{}}
	uc := application.NewFetchWatchedPosts(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: 各 id へ ListByUser が呼ばれる
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("ListByUser calls = %d, want 2: %+v", len(fake.calls), fake.calls)
	}
	if fake.calls[0].userID != "user-a" || fake.calls[1].userID != "user-b" {
		t.Fatalf("userIDs = %q, %q, want user-a, user-b", fake.calls[0].userID, fake.calls[1].userID)
	}
	if got == nil {
		t.Fatal("got = nil, want empty slice")
	}
}

func TestFetchWatchedPosts_passesSinceAsNowMinusFetchWindow_whenNowGiven(t *testing.T) {
	// Given: 固定 now と 1 user
	withWatchUserIDs(t, []string{"user-a"})
	fake := &fakePostSource{}
	uc := application.NewFetchWatchedPosts(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)
	wantSince := now.Add(-constants.FetchWindow)

	// When: Run を呼ぶ
	_, err := uc.Run(context.Background(), now)

	// Then: since は now - FetchWindow
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("ListByUser calls = %d, want 1", len(fake.calls))
	}
	if !fake.calls[0].since.Equal(wantSince) {
		t.Fatalf("since = %v, want %v", fake.calls[0].since, wantSince)
	}
}

func TestFetchWatchedPosts_mergesPostsInOrder_whenMultipleUsersReturnPosts(t *testing.T) {
	// Given: 2 user の投稿がある
	withWatchUserIDs(t, []string{"user-a", "user-b"})
	fake := &fakePostSource{
		posts: map[string][]models.Post{
			"user-a": {{ID: "a1", AuthorID: "user-a"}, {ID: "a2", AuthorID: "user-a"}},
			"user-b": {{ID: "b1", AuthorID: "user-b"}},
		},
	}
	uc := application.NewFetchWatchedPosts(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: WatchUserIDs 順で結合される
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	wantIDs := []string{"a1", "a2", "b1"}
	if len(got) != len(wantIDs) {
		t.Fatalf("len = %d, want %d: %+v", len(got), len(wantIDs), got)
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Fatalf("got[%d].ID = %q, want %q", i, got[i].ID, id)
		}
	}
}

func TestFetchWatchedPosts_returnsFirstErrorWithoutPartialResults_whenFirstUserFails(t *testing.T) {
	// Given: 先頭 user の ListByUser が失敗する
	withWatchUserIDs(t, []string{"user-a", "user-b"})
	boom := errors.New("list failed")
	fake := &fakePostSource{
		posts: map[string][]models.Post{
			"user-a": {{ID: "a1"}},
			"user-b": {{ID: "b1"}},
		},
		errs: map[string]error{"user-a": boom},
	}
	uc := application.NewFetchWatchedPosts(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: 最初の error、calls == 1、部分結果なし
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if got != nil {
		t.Fatalf("got = %+v, want nil", got)
	}
	if len(fake.calls) != 1 {
		t.Fatalf("ListByUser calls = %d, want 1 (fail-fast)", len(fake.calls))
	}
}

func TestFetchWatchedPosts_returnsFirstErrorWithoutPartialResults_whenLaterUserFails(t *testing.T) {
	// Given: 先頭は成功、後続 user の ListByUser が失敗する
	withWatchUserIDs(t, []string{"user-a", "user-b"})
	boom := errors.New("list failed")
	fake := &fakePostSource{
		posts: map[string][]models.Post{
			"user-a": {{ID: "a1"}},
			"user-b": {{ID: "b1"}},
		},
		errs: map[string]error{"user-b": boom},
	}
	uc := application.NewFetchWatchedPosts(fake)
	now := time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

	// When: Run を呼ぶ
	got, err := uc.Run(context.Background(), now)

	// Then: error、calls == 2、部分結果なし
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if got != nil {
		t.Fatalf("got = %+v, want nil (partial results discarded)", got)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("ListByUser calls = %d, want 2", len(fake.calls))
	}
	if fake.calls[0].userID != "user-a" || fake.calls[1].userID != "user-b" {
		t.Fatalf("userIDs = %q, %q, want user-a, user-b", fake.calls[0].userID, fake.calls[1].userID)
	}
}
