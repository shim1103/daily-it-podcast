package hackernews_test

import (
	"testing"

	_ "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
)

// TODO(C): List 実装後に本テストを実装する。
// 現状は stub（List が panic("not implemented")）に対する RED。
// @given topstories/item の double @when List(ctx, since)
// @then SourceItem の SourceID / OccurredAt(UTC>=since) / Context 契約を満たす。

func TestList_mapsTopStoryToSourceItem_whenStoryInWindow(t *testing.T) {
	t.Skip("C: List 未実装（stub）。stubRoundTripper で topstories.json→item/<id>.json の double を組み、" +
		"SourceID=SourceID / OccurredAt=time.Unix(item.time,0).UTC() / Context=title,text,url 行を assert する")
}

func TestList_filtersToTypeStory_whenJobOrPollPresent(t *testing.T) {
	t.Skip("C: 同上。type!=\"story\"（job / poll）の id を double に混ぜ、結果へ現れないことを assert する")
}

func TestList_excludesDeletedOrDeadItems(t *testing.T) {
	t.Skip("C: 同上。deleted:true / dead:true の item を double に混ぜ、結果から除外されることを assert する")
}

func TestList_excludesItemsOlderThanSince_atBoundary(t *testing.T) {
	t.Skip("C: 同上。time == since は含み、time == since-1s は除外する境界を assert する")
}

func TestList_fetchesTopLevelCommentsUpToMaxCommentsPerStory(t *testing.T) {
	t.Skip("C: 同上。kids を MaxCommentsPerStory 超で与え、取得数が上限で頭打ち（CommentDepth=1）になることを assert する")
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	t.Skip("C: 同上。全 story が since より古い double で、戻りが len==0 かつ非 nil slice であることを assert する")
}

func TestList_returnsInfrastructureError_whenClientNilOrNon200OrInvalidJSON(t *testing.T) {
	t.Skip("C: 同上。client nil / 非 200 応答 / 壊れた JSON の各経路で *hackernews.Error が返ることを assert する")
}
