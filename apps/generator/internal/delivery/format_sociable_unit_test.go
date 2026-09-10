package delivery_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/adaptererror"
)

func requireLine(t *testing.T, got, want string) {
	t.Helper()
	for _, line := range strings.Split(got, "\n") {
		if line == want {
			return
		}
	}
	t.Fatalf("Format() = %q, want line %q", got, want)
}

func TestFormat_printsDomainKindAndOp_whenErrorIsDomain(t *testing.T) {
	t.Parallel()

	// Given: Domain Error
	err := domainerrors.DomainErr(domainerrors.OpNoSourceItems, errors.New("items is empty"))

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=domain と Op が出る
	requireLine(t, got, "generator: kind=domain")
	requireLine(t, got, "generator: op="+domainerrors.OpNoSourceItems)
}

func TestFormat_printsConfigKindAndKey_whenErrorIsConfigError(t *testing.T) {
	t.Parallel()

	// Given: 単一 config.Error
	err := &config.Error{Key: config.CursorAPIKeyEnv, Kind: config.KindMissing}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=config と key が出る
	requireLine(t, got, "generator: kind=config")
	requireLine(t, got, "generator: op="+config.CursorAPIKeyEnv)
}

func TestFormat_printsConfigKindWithoutOp_whenErrorIsConfigErrorsBundle(t *testing.T) {
	t.Parallel()

	// Given: 複数 key の config.Errors
	err := &config.Errors{Violations: []*config.Error{
		{Key: config.CursorAPIKeyEnv, Kind: config.KindMissing},
		{Key: config.GeminiAPIKeyEnv, Kind: config.KindEmpty},
	}}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=config であり、束ね key を op に選ばない
	requireLine(t, got, "generator: kind=config")
	if strings.Contains(got, "generator: op=") {
		t.Fatalf("Format() = %q, want no op line for Errors bundle", got)
	}
}

func TestFormat_printsInfrastructureKindAndOp_whenErrorIsAdapterError(t *testing.T) {
	t.Parallel()

	// Given: Driven Adapter の Infrastructure Error（source は代表として gdrive）
	err := adaptererror.New("gdrive", "write", errors.New("quota"))

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure と Op が出る
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=write")
}

func TestFormat_printsInfrastructureKindAndOp_whenFmtWrapsAdapterError(t *testing.T) {
	t.Parallel()

	// Given: fmt wrap の内側に Infrastructure Error（原稿 fallback Adapter 経路の代表）
	inner := adaptererror.New("geminiapi", "http_status", errors.New("status 401"))
	err := fmt.Errorf("compose: %w", inner)

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: unwrap して infrastructure と判定し、Op も出る
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=http_status")
}

func TestFormat_printsUnknownKindWithoutOp_whenErrorIsPlain(t *testing.T) {
	t.Parallel()

	// Given: 分類不能な plain error
	err := errors.New("produce failed")

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=unknown で op 行は無い
	requireLine(t, got, "generator: kind=unknown")
	if strings.Contains(got, "generator: op=") {
		t.Fatalf("Format() = %q, want no op line", got)
	}
}

func TestFormat_printsOuterAdapterOp_whenAdapterErrorWrapsInnerError(t *testing.T) {
	t.Parallel()

	// Given: Infrastructure Error が内側 error を wrap した chain
	inner := adaptererror.New("cursorapi", "run", errors.New("status 500"))
	err := adaptererror.New("cursorapi", "write", inner)

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: 外側の Op を採用する
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=write")
}

func TestFormat_printsDomainKind_whenFmtWrapsDomainError(t *testing.T) {
	t.Parallel()

	// Given: fmt wrap の内側に Domain Error
	inner := domainerrors.DomainErr(domainerrors.OpInvalidManuscriptDraft, errors.New("json"))
	err := fmt.Errorf("compose: %w", inner)

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: unwrap して domain と判定する
	requireLine(t, got, "generator: kind=domain")
	requireLine(t, got, "generator: op="+domainerrors.OpInvalidManuscriptDraft)
}

func TestFormat_oneLinesMessage_whenErrorTextHasNewline(t *testing.T) {
	t.Parallel()

	// Given: 改行を含む Error()
	err := errors.New("first\nsecond")

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: message は1行
	requireLine(t, got, "generator: message=first second")
}

func TestFormat_printsUnwrapChainAsCause_whenErrorWrapsCause(t *testing.T) {
	t.Parallel()

	// Given: Unwrap 可能な Domain Error
	cause := errors.New("items is empty")
	err := domainerrors.DomainErr(domainerrors.OpNoSourceItems, cause)

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: cause[0] に Unwrap 先が出る
	requireLine(t, got, "generator: cause[0]="+cause.Error())
}

func TestFormat_printsEachViolationAsCause_whenErrorIsConfigErrors(t *testing.T) {
	t.Parallel()

	// Given: 複数違反の config.Errors
	first := &config.Error{Key: config.CursorAPIKeyEnv, Kind: config.KindMissing}
	second := &config.Error{Key: config.GeminiAPIKeyEnv, Kind: config.KindEmpty}
	err := &config.Errors{Violations: []*config.Error{first, second}}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: 各違反が cause 行になる
	requireLine(t, got, "generator: cause[0]="+first.Error())
	requireLine(t, got, "generator: cause[1]="+second.Error())
}

func TestFormat_messageEqualsOneLinedError_whenErrorIsDomain(t *testing.T) {
	t.Parallel()

	// Given: Domain Error
	err := domainerrors.DomainErr(domainerrors.OpNoSourceItems, errors.New("items is empty"))

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: message は err.Error() の1行化
	requireLine(t, got, "generator: message="+err.Error())
}

func TestFormat_isNonEmptyWithKind_whenErrorIsNonNil(t *testing.T) {
	t.Parallel()

	// Given: 任意の non-nil error
	err := errors.New("produce failed")

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind 付きで非空
	if got == "" {
		t.Fatal("Format() = empty, want non-empty")
	}
	if !strings.Contains(got, "generator: kind=") {
		t.Fatalf("Format() = %q, want kind=", got)
	}
}
