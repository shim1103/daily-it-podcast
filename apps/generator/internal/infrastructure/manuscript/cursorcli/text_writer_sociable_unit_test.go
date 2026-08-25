package cursorcli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
)

type fakeLauncher struct {
	Commands  []commandlaunch.Command
	CallCount int
	Stdout    []byte
	Err       error
}

func (f *fakeLauncher) Launch(ctx context.Context, command commandlaunch.Command) ([]byte, error) {
	f.CallCount++
	f.Commands = append(f.Commands, command)
	return f.Stdout, f.Err
}

func newTextWriterWithLauncher(stdout []byte, launchErr error) (*TextWriter, *fakeLauncher) {
	fake := &fakeLauncher{Stdout: stdout, Err: launchErr}
	return NewTextWriter(fake), fake
}

func envelope(result string) []byte {
	return []byte(fmt.Sprintf(`{"type":"result","subtype":"success","is_error":false,"result":%q}`, result))
}

func TestWrite_returnsNonEmptyFragment_whenLaunchReturnsValidEnvelope(t *testing.T) {
	t.Parallel()

	// Given: Launcher Stub が result を持つ JSON envelope を返す
	writer, _ := newTextWriterWithLauncher(envelope("本日の IT ニュースをお届けします。"), nil)

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

func TestWrite_passesBriefAsStdinWithoutSecretOrProject_whenLaunchSucceeds(t *testing.T) {
	t.Parallel()

	// Given: Launcher Stub が成功 envelope を返す
	writer, fake := newTextWriterWithLauncher(envelope("断片"), nil)

	// When: brief を渡して Write する
	if _, err := writer.Write(context.Background(), "  導入を書いて  "); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Then: Command は Program・Args・Stdin だけで、secret / project / runtime を持たない
	if fake.CallCount != 1 {
		t.Fatalf("Launch call count = %d, want 1", fake.CallCount)
	}
	cmd := fake.Commands[0]
	if cmd.Program != BinaryName {
		t.Fatalf("Program = %q, want %q", cmd.Program, BinaryName)
	}
	if string(cmd.Stdin) != "導入を書いて" {
		t.Fatalf("Stdin = %q, want trimmed brief", string(cmd.Stdin))
	}
	wantArgs := []string{
		PrintFlag,
		ModeFlag, Mode,
		OutputFlag, OutputFormat,
		ModelFlag, ModelID,
		SandboxFlag, SandboxValue,
		TrustFlag,
	}
	if len(cmd.Args) != len(wantArgs) {
		t.Fatalf("Args = %v, want %v", cmd.Args, wantArgs)
	}
	for i := range wantArgs {
		if cmd.Args[i] != wantArgs[i] {
			t.Fatalf("Args[%d] = %q, want %q (Args = %v)", i, cmd.Args[i], wantArgs[i], cmd.Args)
		}
	}
}

func TestWrite_returnsInfrastructureError_whenStdoutIsNotJSON(t *testing.T) {
	t.Parallel()

	// Given: Launcher Stub が JSON でない stdout を返す
	writer, _ := newTextWriterWithLauncher([]byte("not a json envelope"), nil)

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

	// Given: Launcher Stub が result を持たない envelope を返す
	writer, _ := newTextWriterWithLauncher([]byte(`{"type":"result","subtype":"success"}`), nil)

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

	// Given: Launcher Stub が空白だけの result を返す
	writer, _ := newTextWriterWithLauncher(envelope("   \n  "), nil)

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返る
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
}

func TestWrite_returnsInfrastructureError_whenLaunchExitsNonZero(t *testing.T) {
	t.Parallel()

	// Given: Launcher Stub が非0 exit 相当の error を返す
	launchErr := errors.New("exit status 1")
	writer, _ := newTextWriterWithLauncher(nil, launchErr)

	// When: brief を渡して Write する
	got, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返り、原因 chain が保持される
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
	if !errors.Is(err, launchErr) {
		t.Fatalf("errors.Is(err, launchErr) = false, want true (err = %v)", err)
	}
	if got != "" {
		t.Fatalf("Write() = %q, want empty", got)
	}
}

func TestWrite_wrapsLaunchErrorWithoutAppendingStdin_whenLaunchFails(t *testing.T) {
	t.Parallel()

	// Given: 非0 exit 相当の error を返す Launcher Stub
	const stdinToken = "cursorcli-test-stdin-token"
	fake := &fakeLauncher{Err: errors.New("exit status 1")}
	writer := NewTextWriter(fake)

	// When: 識別可能な brief で Write する
	_, err := writer.Write(context.Background(), stdinToken)

	// Then: Adapter は Launch error を wrap するだけで、stdin を error message へ足さない
	if err == nil {
		t.Fatal("Write() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), stdinToken) {
		t.Fatalf("error message %q contains stdin", err.Error())
	}
}

func TestWrite_returnsInfrastructureError_whenBriefIsBlank(t *testing.T) {
	t.Parallel()

	// Given: Launcher Stub を持つ Adapter
	writer, fake := newTextWriterWithLauncher(envelope("断片"), nil)

	// When: 空白だけの brief で Write する
	_, err := writer.Write(context.Background(), "   \t\n ")

	// Then: Infrastructure Error が返り、Launcher を呼ばない
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
	if fake.CallCount != 0 {
		t.Fatalf("Launch call count = %d, want 0", fake.CallCount)
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

func TestWrite_returnsInfrastructureError_whenLauncherIsNil(t *testing.T) {
	t.Parallel()

	// Given: Launcher が nil の Adapter
	writer := &TextWriter{}

	// When: brief を渡して Write する
	_, err := writer.Write(context.Background(), "導入を書いて")

	// Then: Infrastructure Error が返り、起動しない
	var infra *Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *cursorcli.Error", err, err)
	}
}
