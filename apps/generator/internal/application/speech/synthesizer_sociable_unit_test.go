package speech

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/models"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// Scope: Sociable Unit
// 実物: speech.SpeechSynthesizer（source 切り替え UseCase）
// Double: sources の各要素は port.SpeechSynthesizer の fake Spy。fallback は port.FallbackReporter の Spy。
//
// 振る舞い: 先頭 source が全 texts 分成功したら以降を呼ばない / port.ErrSourceExhausted は
// 部分成功（got）を蓄積し fallback.Fallback を経て残り texts だけ次 source へ渡す /
// 非枯渇 error は蓄積を破棄しそのまま返す / 全 source 使い切りは最後の error を返す /
// source 3 つ以上でも順に切り替わる。

// fakeSpeechSynthesizer は port.SpeechSynthesizer の Spy。呼び出し回数・最後に受け取った texts を記録し、
// 返す audios と error を設定できる。
type fakeSpeechSynthesizer struct {
	got []models.SpeechAudio
	err error

	calls     int
	lastTexts []string
}

func (f *fakeSpeechSynthesizer) SynthesizeAll(_ context.Context, texts []string) ([]models.SpeechAudio, error) {
	f.calls++
	f.lastTexts = texts
	return f.got, f.err
}

var _ port.SpeechSynthesizer = (*fakeSpeechSynthesizer)(nil)

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
func newUCWithSpies(sources ...*fakeSpeechSynthesizer) (*SpeechSynthesizer, *fakeFallback) {
	fallback := &fakeFallback{}
	portSources := make([]port.SpeechSynthesizer, len(sources))
	for i, s := range sources {
		portSources[i] = s
	}
	uc := NewSpeechSynthesizer(portSources, fallback)
	return uc, fallback
}

func audio(label string) models.SpeechAudio {
	return models.SpeechAudio{Content: []byte(label)}
}

func exhaustedErr(inner string) error {
	return fmt.Errorf("%w: %w", port.ErrSourceExhausted, errors.New(inner))
}

// assertAudiosEqual は got が want と同一 audios 列であることを固定する。
func assertAudiosEqual(t *testing.T, got, want []models.SpeechAudio) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SynthesizeAll() audios = %+v, want %+v", got, want)
	}
}

func TestSynthesizeAll_returnsFirstSourceAudios_whenFirstSourceSucceeds(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が texts 全件ぶんの audios を返す
	wantAudios := []models.SpeechAudio{audio("一本目"), audio("二本目")}
	first := &fakeSpeechSynthesizer{got: wantAudios}
	second := &fakeSpeechSynthesizer{got: []models.SpeechAudio{audio("呼ばれないはず")}}
	uc, fallback := newUCWithSpies(first, second)
	texts := []string{"一本目", "二本目"}

	// When: SynthesizeAll する
	got, err := uc.SynthesizeAll(context.Background(), texts)

	// Then: 先頭 source の audios が透過し、2 番目も Fallback も呼ばれない
	if err != nil {
		t.Fatalf("SynthesizeAll() error = %v, want nil", err)
	}
	assertAudiosEqual(t, got, wantAudios)
	if first.calls != 1 {
		t.Fatalf("first calls = %d, want 1", first.calls)
	}
	if !reflect.DeepEqual(first.lastTexts, texts) {
		t.Fatalf("first が受け取った texts = %v, want %v", first.lastTexts, texts)
	}
	if second.calls != 0 {
		t.Fatalf("second calls = %d, want 0", second.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("Fallback calls = %d, want 0", fallback.calls)
	}
}

func TestSynthesizeAll_switchesToSecondSourceWithRemainingTexts_whenFirstSourceExhaustedAfterPartialSuccess(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が 3 本中 1 本だけ合成できてから port.ErrSourceExhausted を返し、
	//        2 番目が残り 2 本ぶんの audios を返す
	firstPartial := []models.SpeechAudio{audio("一本目")}
	first := &fakeSpeechSynthesizer{got: firstPartial, err: exhaustedErr("quota gone")}
	secondAudios := []models.SpeechAudio{audio("二本目"), audio("三本目")}
	second := &fakeSpeechSynthesizer{got: secondAudios}
	uc, fallback := newUCWithSpies(first, second)
	texts := []string{"一本目", "二本目", "三本目"}

	// When: SynthesizeAll する
	got, err := uc.SynthesizeAll(context.Background(), texts)

	// Then: 蓄積した部分成功 + 2 番目の audios が返り、2 番目へは残り texts だけが渡る
	if err != nil {
		t.Fatalf("SynthesizeAll() error = %v, want nil", err)
	}
	assertAudiosEqual(t, got, append(append([]models.SpeechAudio{}, firstPartial...), secondAudios...))
	if fallback.calls != 1 {
		t.Fatalf("Fallback calls = %d, want 1", fallback.calls)
	}
	if fallback.lastEvent != fallbackEventTTSSourceSwitched {
		t.Fatalf("Fallback event = %q, want %q", fallback.lastEvent, fallbackEventTTSSourceSwitched)
	}
	wantRemaining := []string{"二本目", "三本目"}
	if !reflect.DeepEqual(second.lastTexts, wantRemaining) {
		t.Fatalf("second が受け取った texts = %v, want %v", second.lastTexts, wantRemaining)
	}
}

func TestSynthesizeAll_propagatesFirstSourceError_whenErrorIsNotSourceExhausted(t *testing.T) {
	t.Parallel()

	// Given: 先頭 source が部分成功を伴う非枯渇 error を返す
	plain := errors.New("network down")
	first := &fakeSpeechSynthesizer{got: []models.SpeechAudio{audio("捨てられるはず")}, err: plain}
	second := &fakeSpeechSynthesizer{got: []models.SpeechAudio{audio("呼ばれないはず")}}
	uc, fallback := newUCWithSpies(first, second)

	// When: SynthesizeAll する
	got, err := uc.SynthesizeAll(context.Background(), []string{"一本目", "二本目"})

	// Then: 先頭 source の error がそのまま返り、蓄積した部分成功は破棄される。2 番目も Fallback も呼ばれない
	if !errors.Is(err, plain) {
		t.Fatalf("SynthesizeAll() error = %v, want %v", err, plain)
	}
	if got != nil {
		t.Fatalf("SynthesizeAll() audios = %v, want nil（非枯渇 error は部分成功を破棄）", got)
	}
	if second.calls != 0 {
		t.Fatalf("second calls = %d, want 0", second.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("Fallback calls = %d, want 0", fallback.calls)
	}
}

func TestSynthesizeAll_returnsLastSourceError_whenAllSourcesExhausted(t *testing.T) {
	t.Parallel()

	// Given: 先頭・2 番目とも port.ErrSourceExhausted を返す（source は 2 つのみ）
	first := &fakeSpeechSynthesizer{err: exhaustedErr("first boom")}
	secondErr := exhaustedErr("second boom")
	second := &fakeSpeechSynthesizer{err: secondErr}
	uc, fallback := newUCWithSpies(first, second)

	// When: SynthesizeAll する
	got, err := uc.SynthesizeAll(context.Background(), []string{"一本目"})

	// Then: 2 番目は 1 回だけ（3 つ目は無い）、戻りは 2 番目の error
	if second.calls != 1 {
		t.Fatalf("second calls = %d, want 1", second.calls)
	}
	if !errors.Is(err, secondErr) {
		t.Fatalf("SynthesizeAll() error = %v, want 2 番目の error", err)
	}
	if got != nil {
		t.Fatalf("SynthesizeAll() audios = %v, want nil", got)
	}
	if fallback.calls != 2 {
		t.Fatalf("Fallback calls = %d, want 2", fallback.calls)
	}
}

func TestSynthesizeAll_triesThirdSource_whenFirstAndSecondBothExhausted(t *testing.T) {
	t.Parallel()

	// Given: source が 3 つ（可変長）、先頭・2 番目が exhausted、3 番目が成功する
	first := &fakeSpeechSynthesizer{err: exhaustedErr("first boom")}
	second := &fakeSpeechSynthesizer{err: exhaustedErr("second boom")}
	thirdAudios := []models.SpeechAudio{audio("三番目 source の audio")}
	third := &fakeSpeechSynthesizer{got: thirdAudios}
	uc, fallback := newUCWithSpies(first, second, third)

	// When: SynthesizeAll する
	got, err := uc.SynthesizeAll(context.Background(), []string{"一本目"})

	// Then: 3 つとも 1 回ずつ呼ばれ、戻りは 3 番目の audios、Fallback は切り替え回数分（2 回）
	if err != nil {
		t.Fatalf("SynthesizeAll() error = %v, want nil", err)
	}
	assertAudiosEqual(t, got, thirdAudios)
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

func TestNewSpeechSynthesizer_returnsNonNil(t *testing.T) {
	t.Parallel()

	uc := NewSpeechSynthesizer([]port.SpeechSynthesizer{&fakeSpeechSynthesizer{}, &fakeSpeechSynthesizer{}}, &fakeFallback{})
	if uc == nil {
		t.Fatal("NewSpeechSynthesizer が nil を返した")
	}
}
