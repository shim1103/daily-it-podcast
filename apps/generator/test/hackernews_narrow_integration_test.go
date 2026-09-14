// Scope: Narrow Integration
// 実物境界: hackernews.ListItemSource が標準 *http.Client で到達する本番 Hacker News Firebase API
// Double: なし（認証不要の公開 API を実 GET）
// @require 本番 host へ外向き GET できる network。secret は渡さない。
// @ensure List は error なしで非 nil slice を返す（該当なしは空 slice 可）。
// @ensure 各要素の SourceID は hackernews.SourceID。OccurredAt は UTC かつ since 以上。
// @invariant httptest 合成 upstream を使わない。写像 exact・retry 表・件数上限は Sociable Unit の所有。
package test

import (
	"context"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
)

func TestHackerNewsListItemSource_reachesRealBoundary_whenAuthFreeGetSucceeds(t *testing.T) {
	// Given: 本番 API 向け標準 client と、空日 flake を抑える広い since 窓
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	// why: HN は story 最大件数 + top-level comment 取得で HTTP 回数が多く、60s だと切れうる。
	source := hackernews.NewListItemSource(newItemSourceHTTPClient(120 * time.Second))

	// When: 実境界へ List する
	got, err := source.List(context.Background(), since)

	// Then: I/O 契約（成功・非 nil・要素があれば SourceID / OccurredAt）
	assertRealBoundaryItemSourceList(t, got, err, hackernews.SourceID, since)
}
