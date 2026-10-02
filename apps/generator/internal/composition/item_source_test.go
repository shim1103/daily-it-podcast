package composition

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
)

// fixedNow は List 呼び出しへ渡す固定時刻である（application 側 test の慣習に合わせる）。
var fixedNow = time.Date(2024, 12, 10, 15, 0, 0, 0, time.UTC)

// fakeItemSource は port.ItemSource を満たす test double である。
// items をそのまま返し、err が非 nil ならそれを優先して返す。
type fakeItemSource struct {
	items []models.SourceItem
	err   error
}

var _ port.ItemSource = fakeItemSource{}

func (f fakeItemSource) List(_ context.Context, _ time.Time) ([]models.SourceItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func TestCompositeItemSource_concatsListResultsFromAllSources(t *testing.T) {
	t.Parallel()

	// Given: 判別できる SourceItem を返す fake を 2 本
	first := fakeItemSource{items: []models.SourceItem{
		{SourceID: "first-1"},
		{SourceID: "first-2"},
	}}
	second := fakeItemSource{items: []models.SourceItem{
		{SourceID: "second-1"},
	}}

	// When: composite の List を呼ぶ
	got, err := newCompositeItemSource(first, second).List(context.Background(), fixedNow)

	// Then: 全 source の結果が含まれる（連結順序は保証しない）
	if err != nil {
		t.Fatalf("List() が error を返した: %v", err)
	}
	wantIDs := map[string]int{
		"first-1":  1,
		"first-2":  1,
		"second-1": 1,
	}
	if len(got) != len(wantIDs) {
		t.Fatalf("件数 = %d, want %d", len(got), len(wantIDs))
	}
	gotCounts := make(map[string]int, len(got))
	for _, item := range got {
		gotCounts[item.SourceID]++
	}
	for id, wantCount := range wantIDs {
		if gotCounts[id] != wantCount {
			t.Fatalf("SourceID %q の出現回数 = %d, want %d（got=%v）", id, gotCounts[id], wantCount, got)
		}
	}
}

func TestCompositeItemSource_returnsNonNilEmptySlice_whenAllSourcesEmpty(t *testing.T) {
	t.Parallel()

	// Given: いずれも空を返す fake を 2 本
	empties := newCompositeItemSource(
		fakeItemSource{items: []models.SourceItem{}},
		fakeItemSource{items: []models.SourceItem{}},
	)

	// When: composite の List を呼ぶ
	got, err := empties.List(context.Background(), fixedNow)

	// Then: 非 nil の空 slice が返る
	if err != nil {
		t.Fatalf("List() が error を返した: %v", err)
	}
	if got == nil {
		t.Fatal("List() が nil を返した（空 slice であるべき）")
	}
	if len(got) != 0 {
		t.Fatalf("件数 = %d, want 0", len(got))
	}
}

func TestCompositeItemSource_returnsNonNilEmptySlice_whenNoSources(t *testing.T) {
	t.Parallel()

	// Given: source を 1 本も登録しない composite
	empty := newCompositeItemSource()

	// When: composite の List を呼ぶ
	got, err := empty.List(context.Background(), fixedNow)

	// Then: 非 nil の空 slice が返る
	if err != nil {
		t.Fatalf("List() が error を返した: %v", err)
	}
	if got == nil {
		t.Fatal("List() が nil を返した（空 slice であるべき）")
	}
	if len(got) != 0 {
		t.Fatalf("件数 = %d, want 0", len(got))
	}
}

func TestCompositeItemSource_propagatesError_whenAnySourceFails(t *testing.T) {
	t.Parallel()

	// Given: 成功する source と失敗する source が混在する（並行呼び出しでも部分成功を返さないこと）
	sentinel := errors.New("second source failed")
	failing := newCompositeItemSource(
		fakeItemSource{items: []models.SourceItem{{SourceID: "first-1"}}},
		fakeItemSource{err: sentinel},
		fakeItemSource{items: []models.SourceItem{{SourceID: "third-1"}}},
	)

	// When: composite の List を呼ぶ
	got, err := failing.List(context.Background(), fixedNow)

	// Then: その error が返り、成功分は返さない
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want %v", err, sentinel)
	}
	if got != nil {
		t.Fatalf("error 時に成功分が返った: %v", got)
	}
}
