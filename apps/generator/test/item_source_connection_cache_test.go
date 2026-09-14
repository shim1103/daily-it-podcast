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
	runConnectionCacheCase(t, "hackernews", hackernews.SourceID, func(client *http.Client) itemSourceLister {
		return hackernews.NewListItemSource(client)
	})
}

func TestConnectionCache_lobstersListSucceedsAndSaves(t *testing.T) {
	runConnectionCacheCase(t, "lobsters", lobsters.SourceID, func(client *http.Client) itemSourceLister {
		return lobsters.NewListItemSource(client)
	})
}

func TestConnectionCache_publickeyListSucceedsAndSaves(t *testing.T) {
	runConnectionCacheCase(t, "publickey", publickey.SourceID, func(client *http.Client) itemSourceLister {
		return publickey.NewListItemSource(client)
	})
}

func TestConnectionCache_techcrunchListSucceedsAndSaves(t *testing.T) {
	runConnectionCacheCase(t, "techcrunch", techcrunch.SourceID, func(client *http.Client) itemSourceLister {
		return techcrunch.NewListItemSource(client)
	})
}

func TestConnectionCache_cloudwatchListSucceedsAndSaves(t *testing.T) {
	runConnectionCacheCase(t, "cloudwatch", cloudwatch.SourceID, func(client *http.Client) itemSourceLister {
		return cloudwatch.NewListItemSource(client)
	})
}

// itemSourceLister は接続 cache が呼ぶ List 面だけを表す。
type itemSourceLister interface {
	List(ctx context.Context, since time.Time) ([]models.SourceItem, error)
}

func runConnectionCacheCase(t *testing.T, name, wantSourceID string, newSource func(*http.Client) itemSourceLister) {
	t.Helper()

	// Given: 本番向け client と広い since 窓
	// why: HN 含め複数源を同一 helper で回すため、story+comment 多段 fetch でも切れない timeout を共通採用。
	since := time.Now().UTC().Add(-30 * 24 * time.Hour)
	source := newSource(newItemSourceHTTPClient(120 * time.Second))

	// When: 実 HTTP で List し、結果を .cache/ へ保存する
	got, err := source.List(context.Background(), since)

	// Then: I/O 契約は NI と同一 SSoT。cache 保存のみ本 suite 固有。
	assertRealBoundaryItemSourceList(t, got, err, wantSourceID, since)
	saveConnectionCache(t, name, got)
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
