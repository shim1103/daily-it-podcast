package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRun_exitsZeroWithoutStderr_whenProduceReturnsNil(t *testing.T) {
	t.Parallel()

	// Given: ProduceEpisode.Run 相当が nil error を返す
	var stderr bytes.Buffer
	produce := func(context.Context, time.Time) (string, error) {
		return "ep-1", nil
	}

	// When: exit mapping を実行する
	code := run(context.Background(), time.Unix(0, 0).UTC(), &stderr, produce)

	// Then: exit 0 かつ stderr は空
	if code != 0 {
		t.Fatalf("run() code = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRun_exitsNonZeroWithStderr_whenProduceReturnsError(t *testing.T) {
	t.Parallel()

	// Given: ProduceEpisode.Run 相当が non-nil error を返す
	var stderr bytes.Buffer
	wantErr := errors.New("produce failed")
	produce := func(context.Context, time.Time) (string, error) {
		return "", wantErr
	}

	// When: exit mapping を実行する
	code := run(context.Background(), time.Unix(0, 0).UTC(), &stderr, produce)

	// Then: exit 非0 かつ stderr に error が出る
	if code == 0 {
		t.Fatal("run() code = 0, want non-zero")
	}
	got := stderr.String()
	if !strings.Contains(got, wantErr.Error()) {
		t.Fatalf("stderr = %q, want contain %q", got, wantErr.Error())
	}
}
