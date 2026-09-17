package manuscript

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// Scope: Sociable Unit
// 実物: manuscript.TextWriter（source 切り替え UseCase）
// Double: sources の各要素は port.TextWriter の fake Spy。fallback は port.FallbackReporter の Spy。
//
// 振る舞い: 先頭 source 成功で以降を呼ばない / port.ErrSourceExhausted・port.ErrDraftRejected は
// fallback.Fallback を経て次 source へ切り替え、LastAttempt の有無で brief の組み立てが変わる /
// どちらの番兵でもない error は透過し以降を呼ばない / 全 source 使い切りは最後の error を返す /
// source 3 つ以上でも順に切り替わる / brief 空（trim 後）は source を一切呼ばず error。

// fakeTextWriter は port.TextWriter の Spy。呼び出し回数・最後の brief・buildFn 呼出結果を記録し、
// 返す draft と error を設定できる。buildFn を実際に呼ぶかは callBuildFn で制御する。
type fakeTextWriter struct {
	draft       models.ManuscriptDraft
	err         error
	callBuildFn bool

	calls        int
	lastBrief    string
	buildFnDraft models.ManuscriptDraft
	buildFnErr   error
}

func (f *fakeTextWriter) Write(_ context.Context, brief string, buildFn func(string) (models.ManuscriptDraft, error)) (models.ManuscriptDraft, error) {
	f.calls++
	f.lastBrief = brief
	if f.callBuildFn {
		f.buildFnDraft, f.buildFnErr = buildFn(brief)
	}
	return f.draft, f.err
}

// fakeFallback は port.FallbackReporter の Spy。呼び出し回数と最後の event 名を記録する。
type fakeFallback struct {
	calls     int
	lastEvent string
}

func (f *fakeFallback) Fallback(event string) {
	f.calls++
	f.lastEvent = event
}

var _ port.FallbackReporter = (*fakeFallback)(nil)

// newUCWithSpies は fake sources と fallback Spy を束ねた UseCase を返す。
func newUCWithSpies(sources ...*fakeTextWriter) (*TextWriter, *fakeFallback) {
	fallback := &fakeFallback{}
	portSources := make([]port.TextWriter, len(sources))
	for i, s := range sources {
		portSources[i] = s
	}
	uc := NewTextWriter(portSources, fallback)
	return uc, fallback
}

func exhaustedErr(inner string) error {
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, errors.New(inner))
}

func rejectedErr(inner string) error {
	return fmt.Errorf("%w: %w", port.ErrDraftRejected, errors.New(inner))
}

// exhaustedErrWithLastAttempt / rejectedErrWithLastAttempt は、番兵 + port.LastAttempt を
// chain に含めた error を組み立てる。errors.As で取り出せることを呼び出し側が確認できる形。
func exhaustedErrWithLastAttempt(raw string, buildErr error) error {
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, port.LastAttempt{Raw: raw, BuildErr: buildErr})
}

func rejectedErrWithLastAttempt(raw string, buildErr error) error {
	return fmt.Errorf("%w: %w", port.ErrDraftRejected, port.LastAttempt{Raw: raw, BuildErr: buildErr})
}

// noopBuildFn は buildFn の伝播確認に使わない test で埋める Dummy。
func noopBuildFn(string) (models.ManuscriptDraft, error) {
	return models.ManuscriptDraft{}, nil
}

// assertDraftEqual は got が want と同一 draft であることを固定する。
// models.ManuscriptDraft は slice field を持つため == 比較不可であり reflect.DeepEqual を使う。
func assertDraftEqual(t *testing.T, got, want models.ManuscriptDraft) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Write() = %+v, want %+v", got, want)
	}
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

func TestWrite_returnsFirstSourceDraft_whenFirstSourceSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が非空 draft を返す
	first := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "先頭 source の draft"}}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて", noopBuildFn)

	// Then: 先頭 source の draft が透過し、2 番目も Fallback も呼ばれない
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, first.draft)
	if first.calls != 1 {
		t.Fatalf("first calls = %d, want 1", first.calls)
	}
	if second.calls != 0 {
		t.Fatalf("second calls = %d, want 0", second.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("Fallback calls = %d, want 0", fallback.calls)
	}
}

func TestWrite_passesBuildFnThrough_toSource(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が受け取った buildFn を実際に呼ぶ
	first := &fakeTextWriter{callBuildFn: true, draft: models.ManuscriptDraft{Title: "先頭 source の draft"}}
	uc, _ := newUCWithSpies(first)
	wantDraft := models.ManuscriptDraft{Title: "buildFn が返す draft"}
	buildFn := func(raw string) (models.ManuscriptDraft, error) {
		return wantDraft, nil
	}

	// When: Write する
	_, err := uc.Write(context.Background(), "本文の要約から原稿を書いて", buildFn)

	// Then: source が呼んだ buildFn は Write に渡したものと同一であり、その戻り値がそのまま観測できる
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, first.buildFnDraft, wantDraft)
	if first.buildFnErr != nil {
		t.Fatalf("source が観測した buildFn の error = %v, want nil", first.buildFnErr)
	}
}

func TestWrite_switchesToSecondSource_whenFirstSourceReportsSourceExhaustedWithoutLastAttempt(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が LastAttempt を伴わない port.ErrSourceExhausted を返し、2 番目は成功する
	first := &fakeTextWriter{err: exhaustedErr("boom")}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて", noopBuildFn)

	// Then: Fallback 1 回・2 番目 1 回、戻りは 2 番目の draft、brief は「素の brief」のまま渡る
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, second.draft)
	if fallback.calls != 1 {
		t.Fatalf("Fallback calls = %d, want 1", fallback.calls)
	}
	if fallback.lastEvent != fallbackEventSourceSwitched {
		t.Fatalf("Fallback event = %q, want %q", fallback.lastEvent, fallbackEventSourceSwitched)
	}
	if second.calls != 1 {
		t.Fatalf("second calls = %d, want 1", second.calls)
	}
	if second.lastBrief != "本文の要約から原稿を書いて" {
		t.Fatalf("second brief = %q, want 素の brief", second.lastBrief)
	}
}

func TestWrite_switchesToSecondSourceWithRejectionBrief_whenFirstSourceReportsSourceExhaustedWithLastAttempt(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が LastAttempt を伴う port.ErrSourceExhausted を返し、2 番目は成功する
	buildErr := errors.New("invalid draft")
	first := &fakeTextWriter{err: exhaustedErrWithLastAttempt("raw response", buildErr)}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)
	brief := "本文の要約から原稿を書いて"

	// When: Write する
	got, err := uc.Write(context.Background(), brief, noopBuildFn)

	// Then: 2 番目へ渡る brief は port.BuildRejectionBrief で組み立てた rejection brief
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, second.draft)
	if fallback.calls != 1 {
		t.Fatalf("Fallback calls = %d, want 1", fallback.calls)
	}
	wantBrief := port.BuildRejectionBrief(brief, "raw response", buildErr.Error())
	if second.lastBrief != wantBrief {
		t.Fatalf("second brief = %q, want %q", second.lastBrief, wantBrief)
	}
}

func TestWrite_switchesToSecondSourceWithPlainBrief_whenFirstSourceReportsDraftRejectedWithoutLastAttempt(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が LastAttempt を伴わない port.ErrDraftRejected を返し、2 番目は成功する
	firstErr := rejectedErr("invalid after max attempts")
	first := &fakeTextWriter{err: firstErr}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)
	brief := "本文の要約から原稿を書いて"

	// When: Write する
	got, err := uc.Write(context.Background(), brief, noopBuildFn)

	// Then: 2 番目へ渡る brief は port.BuildRejectionBriefWithoutRaw で組み立てた brief
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, second.draft)
	if fallback.calls != 1 {
		t.Fatalf("Fallback calls = %d, want 1", fallback.calls)
	}
	wantBrief := port.BuildRejectionBriefWithoutRaw(brief, firstErr.Error())
	if second.lastBrief != wantBrief {
		t.Fatalf("second brief = %q, want %q", second.lastBrief, wantBrief)
	}
}

func TestWrite_switchesToSecondSourceWithRejectionBrief_whenFirstSourceReportsDraftRejectedWithLastAttempt(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が LastAttempt を伴う port.ErrDraftRejected を返し、2 番目は成功する
	buildErr := errors.New("invalid draft")
	first := &fakeTextWriter{err: rejectedErrWithLastAttempt("raw response", buildErr)}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)
	brief := "本文の要約から原稿を書いて"

	// When: Write する
	got, err := uc.Write(context.Background(), brief, noopBuildFn)

	// Then: 2 番目へ渡る brief は port.BuildRejectionBrief で組み立てた rejection brief
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, second.draft)
	if fallback.calls != 1 {
		t.Fatalf("Fallback calls = %d, want 1", fallback.calls)
	}
	wantBrief := port.BuildRejectionBrief(brief, "raw response", buildErr.Error())
	if second.lastBrief != wantBrief {
		t.Fatalf("second brief = %q, want %q", second.lastBrief, wantBrief)
	}
}

func TestWrite_propagatesFirstSourceError_whenErrorIsNeitherSentinel(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source がどちらの番兵でもない error を返す
	plain := errors.New("plain")
	first := &fakeTextWriter{err: plain}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて", noopBuildFn)

	// Then: 先頭 source の error がそのまま返り、2 番目も Fallback も呼ばれない
	if !errors.Is(err, plain) {
		t.Fatalf("Write() error = %v, want %v", err, plain)
	}
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if second.calls != 0 {
		t.Fatalf("second calls = %d, want 0", second.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("Fallback calls = %d, want 0", fallback.calls)
	}
}

func TestWrite_returnsLastSourceError_whenAllSourcesExhausted(t *testing.T) {
	t.Parallel()

	// Given: 先頭・2 番目とも port.ErrSourceExhausted を返す（source は 2 つのみ）
	first := &fakeTextWriter{err: exhaustedErr("first boom")}
	secondErr := exhaustedErr("second boom")
	second := &fakeTextWriter{err: secondErr}
	uc, fallback := newUCWithSpies(first, second)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて", noopBuildFn)

	// Then: 2 番目は 1 回だけ（3 つ目は無い）、戻りは 2 番目の error
	if second.calls != 1 {
		t.Fatalf("second calls = %d, want 1", second.calls)
	}
	if !errors.Is(err, secondErr) {
		t.Fatalf("Write() error = %v, want 2 番目の error", err)
	}
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	// what: 2 は「先頭 source の exhausted 判定」と「2 番目 source の exhausted 判定」の合計回数。
	if fallback.calls != 2 {
		t.Fatalf("Fallback calls = %d, want 2", fallback.calls)
	}
}

func TestWrite_triesThirdSource_whenFirstAndSecondBothExhausted(t *testing.T) {
	t.Parallel()

	// Given: source が 3 つ（可変長）、先頭・2 番目が exhausted、3 番目が成功する
	first := &fakeTextWriter{err: exhaustedErr("first boom")}
	second := &fakeTextWriter{err: exhaustedErr("second boom")}
	third := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "3 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second, third)

	// When: Write する
	got, err := uc.Write(context.Background(), "本文の要約から原稿を書いて", noopBuildFn)

	// Then: 3 つとも 1 回ずつ呼ばれ、戻りは 3 番目の draft、Fallback は切り替え回数分（2 回）
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	assertDraftEqual(t, got, third.draft)
	if first.calls != 1 {
		t.Fatalf("first calls = %d, want 1", first.calls)
	}
	if second.calls != 1 {
		t.Fatalf("second calls = %d, want 1", second.calls)
	}
	if third.calls != 1 {
		t.Fatalf("third calls = %d, want 1", third.calls)
	}
	if fallback.calls != 2 {
		t.Fatalf("Fallback calls = %d, want 2", fallback.calls)
	}
}

func TestWrite_returnsError_whenBriefEmptyAfterTrim(t *testing.T) {
	t.Parallel()

	// Given: trim 後に空の brief
	first := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "先頭 source の draft"}}
	second := &fakeTextWriter{draft: models.ManuscriptDraft{Title: "2 番目 source の draft"}}
	uc, fallback := newUCWithSpies(first, second)

	// When: Write する
	got, err := uc.Write(context.Background(), "  \t\n  ", noopBuildFn)

	// Then: empty_brief の Domain Error が返り、source も Fallback も一切呼ばれない
	assertDomainOp(t, err, domainerrors.OpEmptyBrief)
	assertDraftEqual(t, got, models.ManuscriptDraft{})
	if first.calls != 0 {
		t.Fatalf("first calls = %d, want 0", first.calls)
	}
	if second.calls != 0 {
		t.Fatalf("second calls = %d, want 0", second.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("Fallback calls = %d, want 0", fallback.calls)
	}
}

func TestNewTextWriter_returnsNonNil(t *testing.T) {
	t.Parallel()

	uc := NewTextWriter([]port.TextWriter{&fakeTextWriter{}, &fakeTextWriter{}}, &fakeFallback{})
	if uc == nil {
		t.Fatal("NewTextWriter が nil を返した")
	}
}
