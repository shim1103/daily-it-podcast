package manuscript

import (
	"context"
	"errors"
	"fmt"
	"testing"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// Scope: Sociable Unit
// 実物: manuscript.TextWriter（切り替え UseCase）
// Double: primary / secondary は port.TextWriter の fake Spy。onFallback は呼び出し回数を数える closure。
//
// 振る舞い: primary 成功で secondary を呼ばない / primary が port.ErrSourceExhausted を wrap したら
// onFallback 1 回 + secondary 1 回で戻りを透過 / 非枯渇 error は透過 / 切り替えは高々 1 回 /
// brief 空（trim 後）は primary を呼ばず error。

// fakeTextWriter は port.TextWriter の Spy。呼び出し回数・最後の ctx/brief を記録し、返す断片と error を設定できる。
type fakeTextWriter struct {
	fragment string
	err      error

	calls     int
	lastCtx   context.Context
	lastBrief string
}

func (f *fakeTextWriter) Write(ctx context.Context, brief string) (string, error) {
	f.calls++
	f.lastCtx = ctx
	f.lastBrief = brief
	return f.fragment, f.err
}

// newUCWithSpies は fake primary / secondary と onFallback カウンタを束ねた UseCase を返す。
func newUCWithSpies(primary, secondary *fakeTextWriter) (*TextWriter, *int) {
	fallbackCalls := 0
	uc := NewTextWriter(primary, secondary, func() { fallbackCalls++ })
	return uc, &fallbackCalls
}

func exhaustedErr(inner string) error {
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, errors.New(inner))
}

// assertDomainOp は err が指定 Op の *domainerrors.Error であることを固定する。
func assertDomainOp(t *testing.T, err error, wantOp string) {
	t.Helper()
	var domErr *domainerrors.Error
	if !errors.As(err, &domErr) {
		t.Fatalf("error type %T (%v), want *domainerrors.Error", err, err)
	}
	if domErr.Op != wantOp {
		t.Fatalf("Op = %q, want %q", domErr.Op, wantOp)
	}
}

func TestWrite_returnsPrimaryFragment_whenPrimarySucceeds(t *testing.T) {
	t.Parallel()

	// Given: primary が非空断片を返す
	primary := &fakeTextWriter{fragment: "primary の原稿断片"}
	secondary := &fakeTextWriter{fragment: "secondary の原稿断片"}
	uc, fallbackCalls := newUCWithSpies(primary, secondary)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: primary 断片が透過し、secondary も onFallback も呼ばれない
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "primary の原稿断片" {
		t.Fatalf("Write() = %q, want %q", got, "primary の原稿断片")
	}
	if primary.calls != 1 {
		t.Fatalf("primary calls = %d, want 1", primary.calls)
	}
	if secondary.calls != 0 {
		t.Fatalf("secondary calls = %d, want 0", secondary.calls)
	}
	if *fallbackCalls != 0 {
		t.Fatalf("onFallback calls = %d, want 0", *fallbackCalls)
	}
}

func TestWrite_switchesToSecondary_whenPrimaryReportsSourceExhausted(t *testing.T) {
	t.Parallel()

	// Given: primary が port.ErrSourceExhausted を wrap した error を返し、secondary は成功する
	primary := &fakeTextWriter{err: exhaustedErr("boom")}
	secondary := &fakeTextWriter{fragment: "secondary の原稿断片"}
	uc, fallbackCalls := newUCWithSpies(primary, secondary)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: onFallback 1 回・secondary 1 回、戻りは secondary の断片、secondary の brief は primary と同一
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "secondary の原稿断片" {
		t.Fatalf("Write() = %q, want %q", got, "secondary の原稿断片")
	}
	if *fallbackCalls != 1 {
		t.Fatalf("onFallback calls = %d, want 1", *fallbackCalls)
	}
	if secondary.calls != 1 {
		t.Fatalf("secondary calls = %d, want 1", secondary.calls)
	}
	if secondary.lastBrief != primary.lastBrief {
		t.Fatalf("secondary brief = %q, want %q (primary と同一)", secondary.lastBrief, primary.lastBrief)
	}
}

func TestWrite_propagatesPrimaryError_whenErrorIsNotSourceExhausted(t *testing.T) {
	t.Parallel()

	// Given: primary が非枯渇 error を返す
	plain := errors.New("plain")
	primary := &fakeTextWriter{err: plain}
	secondary := &fakeTextWriter{fragment: "secondary の原稿断片"}
	uc, fallbackCalls := newUCWithSpies(primary, secondary)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: primary の error がそのまま返り、secondary も onFallback も呼ばれない
	if !errors.Is(err, plain) {
		t.Fatalf("Write() error = %v, want %v", err, plain)
	}
	if got != "" {
		t.Fatalf("Write() fragment = %q, want empty", got)
	}
	if secondary.calls != 0 {
		t.Fatalf("secondary calls = %d, want 0", secondary.calls)
	}
	if *fallbackCalls != 0 {
		t.Fatalf("onFallback calls = %d, want 0", *fallbackCalls)
	}
}

func TestWrite_doesNotCallThirdSource_whenSecondaryAlsoReportsSourceExhausted(t *testing.T) {
	t.Parallel()

	// Given: primary も secondary も port.ErrSourceExhausted を wrap した error を返す
	primary := &fakeTextWriter{err: exhaustedErr("primary boom")}
	secondaryErr := exhaustedErr("secondary boom")
	secondary := &fakeTextWriter{err: secondaryErr}
	uc, fallbackCalls := newUCWithSpies(primary, secondary)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて")

	// Then: secondary は 1 回だけ（3 つ目は無い）、戻りは secondary の error
	if secondary.calls != 1 {
		t.Fatalf("secondary calls = %d, want 1", secondary.calls)
	}
	if !errors.Is(err, secondaryErr) {
		t.Fatalf("Write() error = %v, want secondary の error", err)
	}
	if got != "" {
		t.Fatalf("Write() fragment = %q, want empty", got)
	}
	if *fallbackCalls != 1 {
		t.Fatalf("onFallback calls = %d, want 1", *fallbackCalls)
	}
}

func TestWrite_returnsError_whenBriefEmptyAfterTrim(t *testing.T) {
	t.Parallel()

	// Given: trim 後に空の brief
	primary := &fakeTextWriter{fragment: "primary の原稿断片"}
	secondary := &fakeTextWriter{fragment: "secondary の原稿断片"}
	uc, fallbackCalls := newUCWithSpies(primary, secondary)

	// When: Write する
	got, err := uc.Write(context.Background(), "  \t\n  ")

	// Then: empty_brief の Domain Error が返り、primary も secondary も onFallback も呼ばれない
	assertDomainOp(t, err, domainerrors.OpEmptyBrief)
	if got != "" {
		t.Fatalf("Write() fragment = %q, want empty", got)
	}
	if primary.calls != 0 {
		t.Fatalf("primary calls = %d, want 0", primary.calls)
	}
	if secondary.calls != 0 {
		t.Fatalf("secondary calls = %d, want 0", secondary.calls)
	}
	if *fallbackCalls != 0 {
		t.Fatalf("onFallback calls = %d, want 0", *fallbackCalls)
	}
}

func TestNewTextWriter_returnsNonNil(t *testing.T) {
	t.Parallel()

	uc := NewTextWriter(&fakeTextWriter{}, &fakeTextWriter{}, func() {})
	if uc == nil {
		t.Fatal("NewTextWriter が nil を返した")
	}
}
