package itmedia_test

import (
	"testing"

	_ "github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/itmedia"
)

// TODO(C): List 実装後に本テストを実装する。
// 現状は stub（List が panic("not implemented")）に対する RED。
// @given feedURL の RSS 2.0 応答 double @when List(ctx, since)
// @then SourceItem の SourceID / OccurredAt(UTC>=since) / Context 契約を満たす。

func TestList_parsesFeedAndMapsItemToSourceItem_whenItemInWindow(t *testing.T) {
	t.Skip("C: List 未実装（stub）。stubRoundTripper で feedURL の RSS 2.0 XML を返す double を組み、" +
		"SourceID=SourceID / OccurredAt=pubDate(RFC1123Z).UTC() / Context=title,description,\"links: <link>\" 行を assert する")
}

func TestList_excludesItemsOlderThanSince_atBoundary(t *testing.T) {
	t.Skip("C: 同上。pubDate == since は含み、since より前は除外する境界を assert する")
}

func TestList_buildsTitleOnlySourceItem_whenDescriptionEmpty(t *testing.T) {
	t.Skip("C: 同上。description が空の item で、Context が title のみ（+ links 行）で組まれることを assert する")
}

func TestList_appendsLinksLineFromItemLink(t *testing.T) {
	t.Skip("C: 同上。Context 末尾へ \"links: <item.link>\" 行が付くことを assert する")
}

func TestList_returnsNonNilEmptySlice_whenNothingInWindow(t *testing.T) {
	t.Skip("C: 同上。全 item が since より古い feed で、戻りが len==0 かつ非 nil slice であることを assert する")
}

func TestList_returnsInfrastructureError_whenNon200OrInvalidXML(t *testing.T) {
	t.Skip("C: 同上。非 200 応答 / 壊れた XML の各経路で *itmedia.Error が返ることを assert する")
}
