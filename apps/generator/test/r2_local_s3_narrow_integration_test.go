//go:build r2locals3

// Scope: Narrow Integration（peer 到達）
// 実物境界: wrangler experimental local S3 の HTTP 到達
// Double: 本番 R2 は使わない。本番 Adapter（EpisodeWriter / Lookup）は刺さない。
// Adapter 振る舞いの正は r2_narrow_integration_test.go（httptest）。
//
// @require r2locals3.Start が実 peer を返す。
// @ensure Peer.BaseURL へ HTTP が届き、応答を観測できる（status 値は問わない）。
// @invariant error message に peer credential 実値を載せない（Start 失敗時は sanitize 済み）。
package test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shim1103/daily-it-podcast/apps/generator/test/r2locals3"
)

func TestLocalS3Peer_isReachable_whenStartSucceeds(t *testing.T) {
	// Given: experimental local S3 peer
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	peer, cleanup, err := r2locals3.Start(ctx)
	if err != nil {
		t.Fatalf("r2locals3.Start: %v", err)
	}
	t.Cleanup(cleanup)

	if peer == nil || peer.BaseURL == "" {
		t.Fatal("peer BaseURL が空")
	}

	// When: S3 API root へ薄い HTTP GET する（Adapter は使わない）
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, peer.BaseURL+"/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	res, err := http.DefaultClient.Do(req)

	// Then: 接続でき応答がある（認証なしでも status は何でもよい。接続失敗だけ NG）。Peer 注入値は空でない（到達確認用の契約面）。
	if err != nil {
		t.Fatalf("peer HTTP: %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode < 100 {
		t.Fatalf("status = %d", res.StatusCode)
	}

	for _, v := range []string{peer.AccessKeyID, peer.SecretAccessKey, peer.AccountID, peer.Bucket} {
		if strings.TrimSpace(v) == "" {
			t.Fatal("peer 注入値が空")
		}
	}
}
