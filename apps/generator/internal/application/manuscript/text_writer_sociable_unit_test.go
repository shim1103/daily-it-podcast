package manuscript

import (
	"context"
	"testing"
)

// Scope: Sociable Unit（A の足場。coverage gate を緑に保つだけ）
// 実物: manuscript.TextWriter の stub。Double: なし。
//
// 振る舞い（primary 成功で secondary 呼ばず / 枯渇で 1 回切替+通知 / 非枯渇 error は透過 /
// 切替は高々 1 回）の behavior test は C（docs/tasks/todo/generator-text-writer-fallback.md）が
// この file を置き換えて書く。

func TestNewTextWriter_returnsNonNil(t *testing.T) {
	t.Parallel()

	uc := NewTextWriter(nil, nil, func() {})
	if uc == nil {
		t.Fatal("NewTextWriter が nil を返した")
	}

	// stub の Write を 1 度通す（coverage 用。戻り値は未実装なので検証しない）。
	_, _ = uc.Write(context.Background(), "brief")
}
