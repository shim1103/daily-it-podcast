package delivery_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
)

func TestEvent_writesSingleLineWithCategoryAndEvent_whenNoFields(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: field 無しで Event を書く
	l.Event("progress", "fetch_source_items")

	// Then: category / event だけの 1 行 format
	if got := buf.String(); got != "generator: category=progress event=fetch_source_items\n" {
		t.Fatalf("Event() = %q, want %q", got, "generator: category=progress event=fetch_source_items\n")
	}
}

func TestEvent_appendsFieldsAsKeyValuePairs_whenFieldsGiven(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: fallback 通知を category=fallback で書く
	l.Event("fallback", "manuscript_source_switched",
		delivery.Field{Key: "from", Value: "cursor"},
		delivery.Field{Key: "to", Value: "gemini"},
		delivery.Field{Key: "reason", Value: "source_exhausted"})

	// Then: category event の後ろへ field が key=value で並ぶ 1 行
	want := "generator: category=fallback event=manuscript_source_switched from=cursor to=gemini reason=source_exhausted\n"
	if got := buf.String(); got != want {
		t.Fatalf("Event() = %q, want %q", got, want)
	}
}

func TestEvent_collapsesNewlinesInFieldValue_whenValueHasNewline(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: 改行を含む value を渡す
	l.Event("progress", "step", delivery.Field{Key: "note", Value: "line1\nline2\r\nline3"})

	// Then: 改行は空白へ潰れて出力全体が 1 行（末尾 \n を除く）
	got := buf.String()
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("Event() = %q, want exactly one trailing newline", got)
	}
	if !strings.Contains(got, "note=line1 line2 line3") {
		t.Fatalf("Event() = %q, want collapsed value %q", got, "note=line1 line2 line3")
	}
}

func TestEvent_doesNothing_whenLogWriterPointerIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil の *LogWriter
	var l *delivery.LogWriter

	// When: Event を呼ぶ
	// Then: panic せず no-op（呼び出しが返ればよい）
	l.Event("progress", "step")
}

func TestEvent_doesNothing_whenUnderlyingWriterIsNil(t *testing.T) {
	t.Parallel()

	// Given: writer を渡さない LogWriter
	l := delivery.NewLogWriter(nil)

	// When: Event を呼ぶ
	// Then: panic せず no-op
	l.Event("progress", "step", delivery.Field{Key: "k", Value: "v"})
}

func TestStart_writesCategoryProgressWithPhaseStart(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: 段階開始を書く
	l.Start("fetch_source_items")

	// Then: category=progress phase=start の 1 行
	if got := buf.String(); got != "generator: category=progress event=fetch_source_items phase=start\n" {
		t.Fatalf("Start() = %q, want %q", got, "generator: category=progress event=fetch_source_items phase=start\n")
	}
}

func TestDone_writesCategoryProgressWithPhaseDoneAndDetail_whenDetailGiven(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: detail 付きで段階完了を書く
	l.Done("fetch_source_items", "3件")

	// Then: category=progress phase=done detail=<値>
	if got := buf.String(); got != "generator: category=progress event=fetch_source_items phase=done detail=3件\n" {
		t.Fatalf("Done() = %q, want %q", got, "generator: category=progress event=fetch_source_items phase=done detail=3件\n")
	}
}

func TestDone_omitsDetailField_whenDetailEmpty(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: detail 空で段階完了を書く
	l.Done("compose_brief", "")

	// Then: phase=done だけで detail= を出さない
	if got := buf.String(); got != "generator: category=progress event=compose_brief phase=done\n" {
		t.Fatalf("Done() = %q, want %q", got, "generator: category=progress event=compose_brief phase=done\n")
	}
}

func TestFallback_writesCategoryFallbackWithEventName(t *testing.T) {
	t.Parallel()

	// Given: bytes.Buffer を出力先にした LogWriter
	var buf bytes.Buffer
	l := delivery.NewLogWriter(&buf)

	// When: 切替イベントを書く
	l.Fallback("manuscript_source_switched")

	// Then: category=fallback の 1 行
	if got := buf.String(); got != "generator: category=fallback event=manuscript_source_switched\n" {
		t.Fatalf("Fallback() = %q, want %q", got, "generator: category=fallback event=manuscript_source_switched\n")
	}
}
