//go:build !r2locals3

package r2locals3_test

import (
	"context"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/test/r2locals3"
)

// Fake 契約: 既定 build の Start は wrangler を起動せず zero Peer を返す。
func TestStart_returnsZeroPeerAndNoopCleanup_whenFakeBuild(t *testing.T) {
	t.Parallel()

	// Given: !r2locals3 Fake Start
	// When: Start する
	peer, cleanup, err := r2locals3.Start(context.Background())

	// Then: nil error、非 nil peer、BaseURL 空、cleanup 呼び出し可能
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if peer == nil {
		t.Fatal("peer = nil")
	}
	if peer.BaseURL != "" {
		t.Fatalf("BaseURL = %q, want empty (Fake)", peer.BaseURL)
	}
	if cleanup == nil {
		t.Fatal("cleanup = nil")
	}
	cleanup()
}
