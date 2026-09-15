package r2locals3_test

import (
	"context"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/test/r2locals3"
)

// A 足場: Start の signature / zero 契約だけを固定する。実 peer 起動は C。
func TestStart_returnsZeroPeerAndNoopCleanup_whenStub(t *testing.T) {
	t.Parallel()

	// Given: stub Start
	// When: Start する
	peer, cleanup, err := r2locals3.Start(context.Background())

	// Then: nil error、非 nil peer、cleanup 呼び出し可能
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if peer == nil {
		t.Fatal("peer = nil")
	}
	if cleanup == nil {
		t.Fatal("cleanup = nil")
	}
	cleanup()
}
