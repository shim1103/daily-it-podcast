package delivery_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/delivery"
	domainerrors "github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/errors"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/drive/gdrive"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/google/oauth"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/speech/gemini"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/x/getxapi"
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

func TestFormat_printsInfrastructureKindAndOp_whenErrorIsCursorcli(t *testing.T) {
	t.Parallel()

	// Given: cursorcli Infrastructure Error
	err := &cursorcli.Error{Op: "run", Err: errors.New("exit status 1")}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure と Op が出る
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=run")
}

func TestFormat_printsInfrastructureKind_whenErrorIsProcessenv(t *testing.T) {
	t.Parallel()

	// Given: processenv Infrastructure Error
	err := &processenv.Error{Op: "launch", Err: errors.New("program is empty")}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=launch")
}

func TestFormat_printsInfrastructureKind_whenErrorIsGemini(t *testing.T) {
	t.Parallel()

	// Given: gemini Infrastructure Error
	err := &gemini.Error{Op: "synthesize", Err: errors.New("timeout")}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=synthesize")
}

func TestFormat_printsInfrastructureKind_whenErrorIsGetxapi(t *testing.T) {
	t.Parallel()

	// Given: getxapi Infrastructure Error
	err := &getxapi.Error{Op: "fetch", Err: errors.New("status 503")}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=fetch")
}

func TestFormat_printsInfrastructureKind_whenErrorIsGdrive(t *testing.T) {
	t.Parallel()

	// Given: gdrive Infrastructure Error
	err := &gdrive.Error{Op: "write", Err: errors.New("quota")}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=write")
}

func TestFormat_printsInfrastructureKind_whenErrorIsOauth(t *testing.T) {
	t.Parallel()

	// Given: oauth Infrastructure Error
	err := &oauth.Error{Op: "refresh", Err: errors.New("invalid_grant")}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: kind=infrastructure
	requireLine(t, got, "generator: kind=infrastructure")
	requireLine(t, got, "generator: op=refresh")
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

func TestFormat_printsCursorcliKindNotInnerProcessenv_whenCursorcliWrapsProcessenv(t *testing.T) {
	t.Parallel()

	// Given: cursorcli が processenv を wrap した chain
	inner := &processenv.Error{Op: "run", Err: errors.New("exit status 1")}
	err := &cursorcli.Error{Op: "write", Err: inner}

	// When: External 表現へ写す
	got := delivery.Format(err)

	// Then: 外側の cursorcli Op を採用する
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
