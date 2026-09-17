package port_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
)

// Scope: Sociable Unit（番兵 error の identity と wrap 透過）
// 実物: port.ErrSourceExhausted。Double: なし。

func TestErrSourceExhausted_isMatchedThroughWrap(t *testing.T) {
	t.Parallel()

	// Given: infra 由来の error を ErrSourceExhausted で wrap した error
	infra := errors.New("cursorapi: create_status: create status 401")
	wrapped := fmt.Errorf("%w: %w", port.ErrSourceExhausted, infra)

	// Then: errors.Is で番兵として検出でき、内側の infra error も辿れる
	if !errors.Is(wrapped, port.ErrSourceExhausted) {
		t.Fatal("errors.Is(wrapped, ErrSourceExhausted) が false")
	}
	if !errors.Is(wrapped, infra) {
		t.Fatal("wrap した infra error が errors.Is で辿れない")
	}
}

func TestErrSourceExhausted_doesNotMatchUnrelatedError(t *testing.T) {
	t.Parallel()

	// Given: 無関係な error
	other := errors.New("some other failure")

	// Then: 番兵として検出されない
	if errors.Is(other, port.ErrSourceExhausted) {
		t.Fatal("無関係な error が ErrSourceExhausted と一致した")
	}
}

func TestLastAttempt_Error_returnsBuildErrMessage_whenBuildErrNonNil(t *testing.T) {
	t.Parallel()

	// Given: BuildErr が非 nil な LastAttempt
	attempt := port.LastAttempt{Raw: "raw response", BuildErr: errors.New("invalid draft")}

	// Then: BuildErr のメッセージを含む
	want := "previous attempt rejected: invalid draft"
	if got := attempt.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestLastAttempt_Error_doesNotPanic_whenBuildErrNil(t *testing.T) {
	t.Parallel()

	// Given: BuildErr が nil な zero value の LastAttempt
	attempt := port.LastAttempt{}

	// Then: panic せず固定文言を返す
	want := "previous attempt rejected: (no build error)"
	if got := attempt.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestBuildRejectionBriefWithoutRaw_embedsReasonWithoutRaw(t *testing.T) {
	t.Parallel()

	// Given: brief と rejection 理由
	brief := "本文の要約から原稿を書いて"
	reason := "invalid draft"

	// Then: Prefix/Middle/Suffix で reason を挟む（raw は挟まない）
	want := brief + port.RejectionPrefixText + port.RejectionMiddleText + reason + port.RejectionSuffixText
	if got := port.BuildRejectionBriefWithoutRaw(brief, reason); got != want {
		t.Fatalf("BuildRejectionBriefWithoutRaw() = %q, want %q", got, want)
	}
}
