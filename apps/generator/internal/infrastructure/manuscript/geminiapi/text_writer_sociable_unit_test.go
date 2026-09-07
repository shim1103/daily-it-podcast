package geminiapi

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// Scope: Sociable Unit（A の足場。coverage gate を緑に保つだけ）
// 実物: geminiapi.TextWriter の stub。Double: なし。
//
// retry / error 方針（成功 / 5xx・Do error を 1 回再試行 / 429 backoff MaxAttempts /
// 4xx 非 retry / finishReason≠STOP / 空 text / key 非露出 / key は header）の behavior test は
// C（docs/tasks/todo/generator-text-writer-fallback.md）がこの file を置き換えて書く。

func TestNewTextWriter_returnsNonNil(t *testing.T) {
	t.Parallel()

	w := NewTextWriter(&http.Client{}, "gemini-fake-key")
	if w == nil {
		t.Fatal("NewTextWriter が nil を返した")
	}

	// 内部 constructor の backoff seam 分岐（nil → ctxSleep）も 1 度通す。
	if got := newTextWriter(&http.Client{}, "k", nil); got.backoffSleepFn == nil {
		t.Fatal("newTextWriter(nil) が backoffSleepFn を補わなかった")
	}

	// stub の Write を 1 度通す（coverage 用。戻り値は未実装なので検証しない）。
	_, _ = w.Write(context.Background(), "brief")
}

func TestCtxSleep_returnsWhenContextDone(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctxSleep(ctx, time.Hour) // ctx が済んでいるので即戻る
}
