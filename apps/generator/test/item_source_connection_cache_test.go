//go:build connectioncache

// Scope: 接続 cache suite（Integration gate 外）
// 実物境界: 5 情報源 Adapter が標準 *http.Client で到達する本番 feed / API
// Double: なし。結果を repo 根の .cache/ へ保存するだけ（production Adapter は cache を知らない）
// @require 本番 host へ外向き GET できる network。secret は渡さない。
// @ensure 各源 List 正常系が到達し、非 nil slice を返す。結果 JSON を .cache/ へ書く。
// @invariant build tag connectioncache により test-integration.sh / test-unit.sh の既定収集から除外。
package test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/cloudwatch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/hackernews"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/lobsters"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/publickey"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/techcrunch"
)

func TestConnectionCache_hackernewsListSucceedsAndSaves(t *testing.T) {
	// Given: HackerNews 本番向け client と広い since 窓
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	source := hackernews.NewListItemSource(newConnectionCacheHTTPClient(120*time.Second), hackernews.MaxStoriesScanned, &retryReporterSpy{})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: 本番 I/O 契約を満たし、結果を .cache/ へ保存する
	assertConnectionCacheItemSourceList(t, got, err, hackernews.SourceID, since)
	saveConnectionCache(t, "hackernews", got)
}

func TestConnectionCache_lobstersListSucceedsAndSaves(t *testing.T) {
	// Given: Lobsters 本番向け client と広い since 窓
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	source := lobsters.NewListItemSource(newConnectionCacheHTTPClient(120*time.Second), lobsters.MaxStoriesScanned, &retryReporterSpy{})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: 本番 I/O 契約を満たし、結果を .cache/ へ保存する
	assertConnectionCacheItemSourceList(t, got, err, lobsters.SourceID, since)
	saveConnectionCache(t, "lobsters", got)
}

func TestConnectionCache_publickeyListSucceedsAndSaves(t *testing.T) {
	// Given: Publickey 本番向け client と広い since 窓
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	source := publickey.NewListItemSource(newConnectionCacheHTTPClient(120*time.Second), publickey.MaxStoriesScanned, &retryReporterSpy{})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: 本番 I/O 契約を満たし、結果を .cache/ へ保存する
	assertConnectionCacheItemSourceList(t, got, err, publickey.SourceID, since)
	saveConnectionCache(t, "publickey", got)
}

func TestConnectionCache_techcrunchListSucceedsAndSaves(t *testing.T) {
	// Given: TechCrunch 本番向け client と広い since 窓
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	source := techcrunch.NewListItemSource(newConnectionCacheHTTPClient(120*time.Second), techcrunch.MaxStoriesScanned, &retryReporterSpy{})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: 本番 I/O 契約を満たし、結果を .cache/ へ保存する
	assertConnectionCacheItemSourceList(t, got, err, techcrunch.SourceID, since)
	saveConnectionCache(t, "techcrunch", got)
}

func TestConnectionCache_cloudwatchListSucceedsAndSaves(t *testing.T) {
	// Given: クラウド Watch 本番向け client と広い since 窓
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	source := cloudwatch.NewListItemSource(newConnectionCacheHTTPClient(120*time.Second), cloudwatch.MaxStoriesScanned, &retryReporterSpy{})

	// When: List(ctx, since) を呼ぶ
	got, err := source.List(context.Background(), since)

	// Then: 本番 I/O 契約を満たし、結果を .cache/ へ保存する
	assertConnectionCacheItemSourceList(t, got, err, cloudwatch.SourceID, since)
	saveConnectionCache(t, "cloudwatch", got)
}

// newConnectionCacheHTTPClient は接続 cache suite 専用の標準 *http.Client を返す。
// why: HN 含め複数源を回すため、story+comment 多段 fetch でも切れない timeout を呼び出し側が渡す。
func newConnectionCacheHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// assertConnectionCacheItemSourceList は接続 cache が観測する本番直撃 I/O 契約を assert する。
// 写像 exact・retry 表・件数上限は Sociable Unit の所有。NI（controllable peer）とは別 suite。
func assertConnectionCacheItemSourceList(t *testing.T, got []models.SourceItem, err error, wantSourceID string, since time.Time) {
	t.Helper()
	if err != nil {
		t.Fatalf("List() error = %v, want nil（認証不要 GET が成功すること）", err)
	}
	if got == nil {
		t.Fatal("List() = nil, want non-nil slice（該当なしは空 slice）")
	}
	for i, item := range got {
		if item.SourceID != wantSourceID {
			t.Fatalf("got[%d].SourceID = %q, want %q", i, item.SourceID, wantSourceID)
		}
		if item.OccurredAt.Location() != time.UTC {
			t.Fatalf("got[%d].OccurredAt.Location() = %v, want UTC", i, item.OccurredAt.Location())
		}
		if item.OccurredAt.Before(since) {
			t.Fatalf("got[%d].OccurredAt = %v, want >= since %v", i, item.OccurredAt, since)
		}
	}
}

func saveConnectionCache(t *testing.T, name string, items []models.SourceItem) {
	t.Helper()
	root := connectionCacheRepoRoot(t)
	dir := filepath.Join(root, ".cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir .cache: %v", err)
	}
	path := filepath.Join(dir, "item-source-"+name+".json")
	raw, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write cache %s: %v", path, err)
	}
	t.Logf("saved connection cache: %s (%d items)", path, len(items))
}

func connectionCacheRepoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse --show-toplevel: %v", err)
	}
	return strings.TrimSpace(string(out))
}
