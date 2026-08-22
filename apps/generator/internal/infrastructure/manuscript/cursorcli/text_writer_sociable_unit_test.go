package cursorcli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

type execProbe struct {
	Names     []string
	Args      [][]string
	Dirs      []string
	Envs      [][]string
	Stdins    []string
	CallCount int
}

const testProjectDir = "/path/to/cursor-project"

func newTextWriterWithExec(stdout []byte, execErr error) (*TextWriter, *execProbe) {
	probe := &execProbe{}
	runFn := func(ctx context.Context, name string, args []string, dir string, env []string, stdin string) ([]byte, error) {
		probe.CallCount++
		probe.Names = append(probe.Names, name)
		probe.Args = append(probe.Args, args)
		probe.Dirs = append(probe.Dirs, dir)
		probe.Envs = append(probe.Envs, env)
		probe.Stdins = append(probe.Stdins, stdin)
		return stdout, execErr
	}
	return &TextWriter{wrapper: testWrapper(), runFn: runFn}, probe
}

func envelope(result string) []byte {
	return []byte(fmt.Sprintf(`{"type":"result","subtype":"success","is_error":false,"result":%q}`, result))
}

func TestWrite_returnsNonEmptyFragment_whenExecReturnsValidEnvelope(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が result を持つ JSON envelope を返す
	writer, _ := newTextWriterWithExec(envelope("本日の IT ニュースをお届けします。"), nil)

	// When: brief を渡して Write する
	got, err := writer.Write(context.Background(), "今日の IT ニュースの導入を書いて")

	// Then: envelope の result が非空 text 断片として返る
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if got != "本日の IT ニュースをお届けします。" {
		t.Fatalf("Write() = %q, want envelope result", got)
	}
}

func TestWrite_passesBriefToCursor_whenExecSucceeds(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が成功 envelope を返す
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "  導入を書いて  "); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: trim 済み brief が Cursor へ 1 度だけ渡る
	if probe.CallCount != 1 {
		t.Fatalf("exec call count = %d, want 1", probe.CallCount)
	}
	if probe.Stdins[0] != "導入を書いて" {
		t.Fatalf("stdin = %q, want trimmed brief", probe.Stdins[0])
	}
}

func TestWrite_buildsArgvInFixedOrderWithoutOmission_whenExecSucceeds(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が成功 envelope を返す
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "導入を書いて"); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: exec する program は wrapper の binary であり、argv は
	// env subcommand と -- separator を経て Cursor CLI と flags へ続く固定順序になる。
	if probe.Names[0] != agentsecrets.EnvBinary {
		t.Fatalf("binary = %q, want %q", probe.Names[0], agentsecrets.EnvBinary)
	}
	want := []string{
		agentsecrets.EnvSubcommand,
		agentsecrets.ArgSeparator,
		BinaryName,
		PrintFlag,
		ModeFlag, Mode,
		OutputFlag, OutputFormat,
		ModelFlag, ModelID,
		SandboxFlag, SandboxValue,
		TrustFlag,
	}
	got := probe.Args[0]
	if len(got) != len(want) {
		t.Fatalf("argv = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argv[%d] = %q, want %q (argv = %v)", i, got[i], want[i], got)
		}
	}
}

func TestWrite_passesExplicitMinimalEnv_whenExecSucceeds(t *testing.T) {
	// why: t.Setenv は process 全体の env を書き換えるため t.Parallel と併用できない。

	// Given: 親 process に、Cursor CLI の実行が必要としない秘密が存在する
	t.Setenv("CURSORCLI_TEST_OTHER_VENDOR_SECRET", "other-vendor-secret-value")
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "導入を書いて"); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: 子 process の env は明示構築された最小集合であり、親環境の全継承ではない
	got := probe.Envs[0]
	if got == nil {
		t.Fatal("env = nil, want explicitly built minimal env (nil delegates to parent environ)")
	}
	// why: 親が定義していない名前は entry ごと落とすため、件数は環境で変わる。
	// 上限だけを固定し、下限は「宣言に無い名前が混ざらない」ことで担保する。
	if len(got) > len(minimalEnvNames) {
		t.Fatalf("env = %v, want at most %d entries built from %v", got, len(minimalEnvNames), minimalEnvNames)
	}
	for _, entry := range got {
		name, _, found := strings.Cut(entry, "=")
		if !found {
			t.Fatalf("env entry %q is not NAME=VALUE", entry)
		}
		if !slices.Contains(minimalEnvNames, name) {
			t.Fatalf("env contains %q, want only %v", name, minimalEnvNames)
		}
	}
}

func TestWrite_omitsUnrelatedParentEnv_whenParentHoldsOtherVendorSecret(t *testing.T) {
	// why: t.Setenv は process 全体の env を書き換えるため t.Parallel と併用できない。

	// Given: 親 process に他 vendor の秘密が env として存在する
	const otherVendorSecret = "cursorcli-test-other-vendor-secret-token"
	t.Setenv("CURSORCLI_TEST_OTHER_VENDOR_SECRET", otherVendorSecret)
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "導入を書いて"); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: その秘密は子 process の env にも argv にも現れない
	for _, entry := range probe.Envs[0] {
		if strings.Contains(entry, otherVendorSecret) {
			t.Fatalf("env entry %q leaks unrelated parent secret", entry)
		}
	}
	if joined := strings.Join(probe.Args[0], " "); strings.Contains(joined, otherVendorSecret) {
		t.Fatalf("argv %q leaks unrelated parent secret", joined)
	}
}

func TestBuildMinimalEnv_returnsNonNil_whenParentDefinesNoneOfTheNames(t *testing.T) {
	// why: t.Setenv は process 全体の env を書き換えるため t.Parallel と併用できない。

	// Given: 最小集合の名前がどれも親環境に存在しない
	for _, name := range minimalEnvNames {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatalf("os.Unsetenv() error = %v, want nil", err)
		}
	}

	// When: 最小 env を構築する
	got := buildMinimalEnv()

	// Then: 空でも非 nil を返す。nil は exec で親環境の全継承を意味するため区別が要る。
	if got == nil {
		t.Fatal("env = nil, want non-nil empty slice (nil means parent inheritance at exec)")
	}
	if len(got) != 0 {
		t.Fatalf("env = %v, want empty", got)
	}
}

func TestBuildMinimalEnv_keepsOnlyDefinedNamesWithParentValues_whenParentHasExtraEntries(t *testing.T) {
	// why: t.Setenv は process 全体の env を書き換えるため t.Parallel と併用できない。

	// Given: 最小集合の名前のうち 1 つと、集合外の名前が親環境に存在する
	t.Setenv(minimalEnvNames[0], "parent-value")
	t.Setenv("CURSORCLI_TEST_OUT_OF_SET", "out-of-set-value")

	// When: 最小 env を構築する
	got := buildMinimalEnv()

	// Then: 集合内の名前は親の値を保ったまま残り、集合外は落ちる
	if !slices.Contains(got, minimalEnvNames[0]+"=parent-value") {
		t.Fatalf("env = %v, want to contain %q", got, minimalEnvNames[0]+"=parent-value")
	}
	for _, entry := range got {
		if strings.HasPrefix(entry, "CURSORCLI_TEST_OUT_OF_SET=") {
			t.Fatalf("env = %v, want no out-of-set entry", got)
		}
	}
}

func TestBuildMinimalEnv_omitsName_whenParentDoesNotDefineIt(t *testing.T) {
	// why: t.Setenv は process 全体の env を書き換えるため t.Parallel と併用できない。

	// Given: 最小集合の名前の 1 つが親環境に存在しない
	t.Setenv(minimalEnvNames[0], "")
	if err := os.Unsetenv(minimalEnvNames[0]); err != nil {
		t.Fatalf("os.Unsetenv() error = %v, want nil", err)
	}

	// When: 最小 env を構築する
	got := buildMinimalEnv()

	// Then: 未定義の名前は空値で載せず、entry ごと落とす
	for _, entry := range got {
		if strings.HasPrefix(entry, minimalEnvNames[0]+"=") {
			t.Fatalf("env = %v, want %q to be omitted when undefined", got, minimalEnvNames[0])
		}
	}
}

func TestWrite_returnsInfrastructureError_whenStdoutIsNotJSON(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が JSON でない stdout を返す
	writer, _ := newTextWriterWithExec([]byte("not a json envelope"), nil)

	// When: brief を渡して Write する
	got, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返り、断片は空
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
	if got != "" {
		t.Fatalf("Write() = %q, want empty", got)
	}
}

func TestWrite_returnsInfrastructureError_whenResultFieldIsMissing(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が result を持たない envelope を返す
	writer, _ := newTextWriterWithExec([]byte(`{"type":"result","subtype":"success"}`), nil)

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返る
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenResultIsBlank(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が空白だけの result を返す
	writer, _ := newTextWriterWithExec(envelope("   \n  "), nil)

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返る
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenExecExitsNonZero(t *testing.T) {
	t.Parallel()

	// Given: exec Stub が非0 exit 相当の error を返す
	execErr := errors.New("exit status 1")
	writer, _ := newTextWriterWithExec(nil, execErr)

	// When: brief を渡して Write する
	got, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返り、原因 chain が保持される
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
	if !errors.Is(err, execErr) {
		t.Fatalf("errors.Is(err, execErr) = false, want true (err = %v)", err)
	}
	if got != "" {
		t.Fatalf("Write() = %q, want empty", got)
	}
}

func TestWrite_doesNotExposeStderrText_whenExecFails(t *testing.T) {
	t.Parallel()

	// Given: stderr へ診断文言を書く exec Stub が非0 exit 相当の error を返す
	const stderrText = "cursor-internal-diagnostic-token"
	var stderrSink strings.Builder
	runFn := func(ctx context.Context, name string, args []string, dir string, env []string, stdin string) ([]byte, error) {
		stderrSink.WriteString(stderrText)
		return nil, errors.New("exit status 1")
	}
	writer := &TextWriter{wrapper: testWrapper(), runFn: runFn}

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: stderr は発生済みだが、上位が読む message へ内容が写らない
	if stderrSink.String() != stderrText {
		t.Fatalf("stderr sink = %q, want fixture to have emitted stderr", stderrSink.String())
	}
	if strings.Contains(err.Error(), stderrText) {
		t.Fatalf("error message %q contains stderr text", err.Error())
	}
}

func TestWrite_returnsInfrastructureError_whenBriefIsBlank(t *testing.T) {
	t.Parallel()

	// Given: exec Stub を持つ Adapter
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: 空白だけの brief で Write する
	_, err := writer.Write(context.Background(), "   \t\n ")

	// Then: Infrastructure Error が返り、Cursor を呼ばない
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
	if probe.CallCount != 0 {
		t.Fatalf("exec call count = %d, want 0", probe.CallCount)
	}
}

func TestWrite_returnsInfrastructureError_whenReceiverIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil の Adapter
	var writer *TextWriter

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: panic せず Infrastructure Error が返る
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
}

func TestNewTextWriter_wiresRunnerThatLaunchesGivenProgram_whenConstructedForProduction(t *testing.T) {
	t.Parallel()

	// Given: production factory で組み立てた Adapter
	writer := NewTextWriter()

	// When: 結線された runner を、決定的な POSIX command で直接実行する
	got, err := writer.runFn(context.Background(), "cat", nil, "", nil, "envelope-body")

	// Then: 実 process 起動へ到達し、stdin が stdout へ抜ける
	if err != nil {
		t.Fatalf("runFn() error = %v, want nil", err)
	}
	if string(got) != "envelope-body" {
		t.Fatalf("runFn() = %q, want stdin echoed to stdout", string(got))
	}
}

func TestRunCursorCLI_returnsStdout_whenCommandSucceeds(t *testing.T) {
	t.Parallel()

	// Given: stdin をそのまま stdout へ流す決定的な command
	// When: runner を実行する
	got, err := runCursorCLI(context.Background(), "cat", nil, "", nil, "envelope-body")

	// Then: stdout が戻り、error は無い
	if err != nil {
		t.Fatalf("runCursorCLI() error = %v, want nil", err)
	}
	if string(got) != "envelope-body" {
		t.Fatalf("runCursorCLI() = %q, want stdin echoed to stdout", string(got))
	}
}

func TestRunCursorCLI_returnsErrorWithoutStderrText_whenCommandExitsNonZero(t *testing.T) {
	t.Parallel()

	// Given: stderr へ書いてから非0 exit する決定的な command
	const stderrText = "cursor-internal-diagnostic-token"

	// When: runner を実行する
	got, err := runCursorCLI(context.Background(), "sh", []string{"-c", "printf %s " + stderrText + " 1>&2; exit 1"}, "", nil, "")

	// Then: error が返り、stderr 内容は error へ写らない
	if err == nil {
		t.Fatal("runCursorCLI() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), stderrText) {
		t.Fatalf("error message %q contains stderr text", err.Error())
	}
	if got != nil {
		t.Fatalf("runCursorCLI() = %q, want nil on failure", string(got))
	}
}

func TestRunCursorCLI_appliesGivenEnvOnly_whenCommandPrintsEnviron(t *testing.T) {
	// why: t.Setenv は process 全体の env を書き換えるため t.Parallel と併用できない。

	// Given: 親 process だけが持つ env と、子 process へ渡す明示 env
	t.Setenv("CURSORCLI_TEST_PARENT_ONLY", "parent-only-value")

	// When: environ を stdout へ出す決定的な command を、明示 env 付きで実行する
	got, err := runCursorCLI(
		context.Background(),
		"sh",
		[]string{"-c", "env"},
		"",
		[]string{"CURSORCLI_TEST_EXPLICIT=explicit-value"},
		"",
	)

	// Then: 明示した env だけが子 process に載り、親固有の env は載らない
	if err != nil {
		t.Fatalf("runCursorCLI() error = %v, want nil", err)
	}
	if !strings.Contains(string(got), "CURSORCLI_TEST_EXPLICIT=explicit-value") {
		t.Fatalf("child environ = %q, want explicit entry", string(got))
	}
	if strings.Contains(string(got), "CURSORCLI_TEST_PARENT_ONLY") {
		t.Fatalf("child environ = %q, want no parent-only entry", string(got))
	}
}

func TestRunCursorCLI_returnsError_whenContextIsAlreadyCanceled(t *testing.T) {
	t.Parallel()

	// Given: 実行前に cancel 済みの ctx と、成功するはずの決定的な command
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When: cancel 済み ctx で runner を実行する
	got, err := runCursorCLI(ctx, "cat", nil, "", nil, "envelope-body")

	// Then: process へ渡らず ctx.Err() 由来の error が即座に返る
	if err == nil {
		t.Fatal("runCursorCLI() error = nil, want non-nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("errors.Is(err, context.Canceled) = false, want true (err = %v)", err)
	}
	if got != nil {
		t.Fatalf("runCursorCLI() = %q, want nil on canceled context", string(got))
	}
}

func testWrapper() agentsecrets.EnvWrapper {
	return agentsecrets.EnvWrapper{
		ProjectDir: testProjectDir,
		SecretKeys: []string{secretnames.CursorAPIKeyName},
	}
}

func TestWrite_runsWrapperInCursorProjectDir_whenExecSucceeds(t *testing.T) {
	t.Parallel()

	// Given: Cursor 専用 project の設定 dir を持つ Adapter
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "導入を書いて"); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: wrapper は その dir を working directory として起動する。
	// active project が cwd で決まるため、これが子 process へ渡る秘密の範囲を決める。
	if probe.Dirs[0] != testProjectDir {
		t.Fatalf("dir = %q, want %q", probe.Dirs[0], testProjectDir)
	}
}

func TestWrite_returnsInfrastructureError_whenProjectDirIsBlank(t *testing.T) {
	t.Parallel()

	// Given: project dir を持たない wrapper で組み立てた Adapter
	probe := &execProbe{}
	runFn := func(ctx context.Context, name string, args []string, dir string, env []string, stdin string) ([]byte, error) {
		probe.CallCount++
		return envelope("断片"), nil
	}
	writer := &TextWriter{wrapper: agentsecrets.EnvWrapper{ProjectDir: "  "}, runFn: runFn}

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返り、Cursor CLI を起動しない。
	// why: 絞り込みが外れた状態で起動すると、cwd の project の全 secret が Cursor へ渡る。
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
	if probe.CallCount != 0 {
		t.Fatalf("exec call count = %d, want 0", probe.CallCount)
	}
}

func TestWrite_declaresCursorAPIKeyAsRequiredSecret_whenConstructedForProduction(t *testing.T) {
	t.Parallel()

	// Given: production factory で組み立てた Adapter
	writer := NewTextWriter()

	// When: 依存する秘密の宣言を読む
	got := writer.RequiredSecretKeys()

	// Then: Cursor API key 名だけが宣言される
	if !slices.Equal(got, []string{secretnames.CursorAPIKeyName}) {
		t.Fatalf("RequiredSecretKeys() = %v, want %v", got, []string{secretnames.CursorAPIKeyName})
	}
}

func TestWrite_keepsDeclaredSecretKeyOutOfArgvAndEnv_whenExecSucceeds(t *testing.T) {
	t.Parallel()

	// Given: Cursor API key 名を宣言した Adapter
	writer, probe := newTextWriterWithExec(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "導入を書いて"); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: 宣言は argv にも env にも現れない。wrapper に key 指定機構が無いため、
	// 宣言は契約 documentation としてだけ働き、実行時の入力にはならない。
	if joined := strings.Join(probe.Args[0], " "); strings.Contains(joined, secretnames.CursorAPIKeyName) {
		t.Fatalf("argv %q contains declared secret key name", joined)
	}
	for _, entry := range probe.Envs[0] {
		if strings.Contains(entry, secretnames.CursorAPIKeyName) {
			t.Fatalf("env entry %q contains declared secret key name", entry)
		}
	}
}

func TestRunCursorCLI_startsCommandInGivenDir_whenDirIsSet(t *testing.T) {
	t.Parallel()

	// Given: 存在が保証された dir と、cwd を stdout へ出す決定的な command
	dir := t.TempDir()

	// When: その dir を指定して runner を実行する
	got, err := runCursorCLI(context.Background(), "sh", []string{"-c", "pwd -P"}, dir, nil, "")

	// Then: 子 process の cwd が指定した dir になる
	if err != nil {
		t.Fatalf("runCursorCLI() error = %v, want nil", err)
	}
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() error = %v, want nil", err)
	}
	if strings.TrimSpace(string(got)) != resolved {
		t.Fatalf("child cwd = %q, want %q", strings.TrimSpace(string(got)), resolved)
	}
}

func TestRunCursorCLI_returnsError_whenDirDoesNotExist(t *testing.T) {
	t.Parallel()

	// Given: 存在しない dir
	dir := filepath.Join(t.TempDir(), "absent-project")

	// When: その dir を指定して runner を実行する
	got, err := runCursorCLI(context.Background(), "cat", nil, dir, nil, "envelope-body")

	// Then: 実行時 error として返る。事前検査ではなく起動失敗として畳む。
	if err == nil {
		t.Fatal("runCursorCLI() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("runCursorCLI() = %q, want nil on failure", string(got))
	}
}

func TestNewTextWriter_resolvesCursorProjectDir_whenConstructedForProduction(t *testing.T) {
	// Given: HOME が設定された環境
	t.Setenv("HOME", "/Users/example")

	// When: production factory で組み立てる
	writer := NewTextWriter()

	// Then: Cursor 専用 project の設定 dir が wrapper へ結線される
	want := agentsecrets.DefaultProjectDir(projectName)
	if writer.wrapper.ProjectDir != want {
		t.Fatalf("wrapper.ProjectDir = %q, want %q", writer.wrapper.ProjectDir, want)
	}
	if err := writer.wrapper.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}
