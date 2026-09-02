package lobsters_test

import (
	"testing"

	_ "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/lobsters"
)

// TODO(C): List 実装後に本テストを実装する。
// 現状は stub（List が panic("not implemented")）に対する RED。
// @given hottest/story の double @when List(ctx, since)
// @then SourceItem の SourceID / OccurredAt(UTC>=since) / Context 契約を満たす。

func TestList_mapsHottestStoryToSourceItem_whenStoryInWindow(t *testing.T) {
	t.Skip("C: List 未実装（stub）。stubRoundTripper で hottest.json→/s/<short_id>.json の double を組み、" +
		"SourceID=SourceID / OccurredAt=created_at.UTC() / Context=title,url 行を assert する")
}

func TestList_excludesStoriesOlderThanSince_atBoundary(t *testing.T) {
	t.Skip("C: 同上。created_at == since は含み、since より前は除外する境界を assert する")
}

func TestList_excludesDeletedOrModeratedComments(t *testing.T) {
	t.Skip("C: 同上。deleted / moderated な comment を double に混ぜ、本文へ現れないことを assert する")
}

func TestList_usesCommentPlainForCommentBody(t *testing.T) {
	t.Skip("C: 同上。comment_plain を本文に使い（HTML の comment field は使わない）、MaxCommentsPerStory で頭打ちすることを assert する")
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	t.Skip("C: 同上。全 story が since より古い double で、戻りが len==0 かつ非 nil slice であることを assert する")
}

func TestList_returnsInfrastructureError_whenClientNilOrNon200OrInvalidJSON(t *testing.T) {
	t.Skip("C: 同上。client nil / 非 200 応答 / 壊れた JSON の各経路で *lobsters.Error が返ることを assert する")
}
