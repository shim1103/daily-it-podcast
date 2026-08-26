package agentsecrets_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

const (
	testAllowPATH = "AGENTSECRETS_LAUNCHER_TEST_PATH"
	testAllowHOME = "AGENTSECRETS_LAUNCHER_TEST_HOME"
)

// setupFakeAgentsecretsOnPATH は `agentsecrets env --` をそのまま exec する fake を PATH 先頭へ置く。
// 戻りの projectDir は絶対 path の dummy Cursor project。
func setupFakeAgentsecretsOnPATH(t *testing.T) (projectDir string) {
	t.Helper()
	projectDir = filepath.Join(t.TempDir(), "cursor-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(bin) error = %v, want nil", err)
	}
	script := "#!/bin/sh\nset -eu\n[ \"$1\" = env ] && [ \"$2\" = -- ]\nshift 2\nexec \"$@\"\n"
	fake := filepath.Join(binDir, agentsecrets.EnvBinary)
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v, want nil", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return projectDir
}

func TestLaunch_wrapsChildWithEnvWrapperAndProjectDir_whenRunnerIsInjected(t *testing.T) {
	t.Parallel()

	// Given: 絶対 path の dummy project と、起動内容を記録する runner
	projectDir := filepath.Join(t.TempDir(), "cursor-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	var gotName string
	var gotArgs []string
	var gotDir string
	var gotEnv []string
	var gotStdin []byte
	runner := func(
		_ context.Context,
		name string,
		args []string,
		dir string,
		env []string,
		stdin []byte,
	) ([]byte, error) {
		gotName = name
		gotArgs = slices.Clone(args)
		gotDir = dir
		gotEnv = slices.Clone(env)
		gotStdin = slices.Clone(stdin)
		return []byte(`{"result":"ok"}`), nil
	}
	launcher := agentsecrets.NewLauncher(
		projectDir,
		[]string{secretnames.CursorAPIKeyName},
		[]string{testAllowPATH, testAllowHOME},
		func(key string) (string, bool) {
			switch key {
			case testAllowPATH:
				return "/agentsecrets-launcher-test-bin", true
			case testAllowHOME:
				return "/agentsecrets-launcher-test-home", true
			default:
				return "", false
			}
		},
		runner,
	)

	// When: Cursor 相当の command を Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "agent",
		Args:    []string{"-p", "--trust"},
		Stdin:   []byte("brief-body"),
	})

	// Then: agentsecrets env -- で包み、cwd は project、env は allowlist のみ、秘密値は Go 側 env に無い
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	if string(got) != `{"result":"ok"}` {
		t.Fatalf("Launch() = %q, want envelope stdout", string(got))
	}
	if gotName != agentsecrets.EnvBinary {
		t.Fatalf("runner name = %q, want %q", gotName, agentsecrets.EnvBinary)
	}
	wantArgs := []string{agentsecrets.EnvSubcommand, agentsecrets.ArgSeparator, "agent", "-p", "--trust"}
	if !slices.Equal(gotArgs, wantArgs) {
		t.Fatalf("runner args = %v, want %v", gotArgs, wantArgs)
	}
	if gotDir != projectDir {
		t.Fatalf("runner dir = %q, want projectDir %q", gotDir, projectDir)
	}
	wantEnv := []string{
		testAllowPATH + "=/agentsecrets-launcher-test-bin",
		testAllowHOME + "=/agentsecrets-launcher-test-home",
	}
	if !slices.Equal(gotEnv, wantEnv) {
		t.Fatalf("runner env = %v, want allowlist only %v", gotEnv, wantEnv)
	}
	if string(gotStdin) != "brief-body" {
		t.Fatalf("runner stdin = %q, want brief-body", string(gotStdin))
	}
	joinedEnv := strings.Join(gotEnv, "\n")
	if strings.Contains(joinedEnv, secretnames.CursorAPIKeyName) {
		t.Fatalf("runner env %q contains secret name; Go must not inject secret values", joinedEnv)
	}
}

func TestLaunch_omitsUndefinedAllowlistName_whenLookupMisses(t *testing.T) {
	t.Parallel()

	// Given: allowlist の 1 名だけが lookup で見つかる
	projectDir := filepath.Join(t.TempDir(), "cursor-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	var gotEnv []string
	runner := func(
		_ context.Context,
		_ string,
		_ []string,
		_ string,
		env []string,
		_ []byte,
	) ([]byte, error) {
		gotEnv = slices.Clone(env)
		return []byte(`{"result":"ok"}`), nil
	}
	launcher := agentsecrets.NewLauncher(
		projectDir,
		[]string{secretnames.CursorAPIKeyName},
		[]string{"PATH", "HOME"},
		func(key string) (string, bool) {
			if key == "PATH" {
				return "/agentsecrets-launcher-test-bin", true
			}
			return "", false
		},
		runner,
	)

	// When: Launch する
	_, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "agent"})

	// Then: 未定義名は落ち、定義済みだけが残る。secret は載せない
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	wantEnv := []string{"PATH=/agentsecrets-launcher-test-bin"}
	if !slices.Equal(gotEnv, wantEnv) {
		t.Fatalf("runner env = %v, want %v", gotEnv, wantEnv)
	}
	joinedEnv := strings.Join(gotEnv, "\n")
	if strings.Contains(joinedEnv, secretnames.CursorAPIKeyName) {
		t.Fatalf("runner env %q contains secret name; Go must not inject secret values", joinedEnv)
	}
}

func TestLaunch_failsBeforeChildStart_whenProjectDirIsRelative(t *testing.T) {
	t.Parallel()

	// Given: 相対 path の project dir と、起動すると失敗させるべき runner
	started := false
	runner := func(
		context.Context, string, []string, string, []string, []byte,
	) ([]byte, error) {
		started = true
		return nil, errors.New("runner must not start")
	}
	launcher := agentsecrets.NewLauncher(
		"relative/cursor",
		[]string{secretnames.CursorAPIKeyName},
		nil,
		nil,
		runner,
	)

	// When: Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "agent"})

	// Then: 起動前失敗。runner は呼ばれない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
	if started {
		t.Fatal("runner started, want validate failure before start")
	}
	var infra *agentsecrets.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *agentsecrets.Error", err, err)
	}
}

func TestLaunch_failsBeforeChildStart_whenProgramIsEmpty(t *testing.T) {
	t.Parallel()

	// Given: 妥当な project dir
	projectDir := filepath.Join(t.TempDir(), "cursor-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	started := false
	runner := func(
		context.Context, string, []string, string, []string, []byte,
	) ([]byte, error) {
		started = true
		return nil, errors.New("runner must not start")
	}
	launcher := agentsecrets.NewLauncher(projectDir, nil, nil, nil, runner)

	// When: 空 Program で Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "  "})

	// Then: 起動前失敗
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
	if started {
		t.Fatal("runner started, want empty program rejection")
	}
}

func TestLaunch_errorOmitsStdin_whenRunnerFails(t *testing.T) {
	t.Parallel()

	// Given: stdin token を持ち、exit 由来の短い cause で失敗する runner
	projectDir := filepath.Join(t.TempDir(), "cursor-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	const stdinToken = "agentsecrets-launcher-stdin-token"
	launcher := agentsecrets.NewLauncher(
		projectDir,
		nil,
		nil,
		nil,
		func(context.Context, string, []string, string, []string, []byte) ([]byte, error) {
			return nil, errors.New("exit status 1")
		},
	)

	// When: Launch する
	_, err := launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "agent",
		Stdin:   []byte(stdinToken),
	})

	// Then: error は返るが stdin token は message に出ない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), stdinToken) {
		t.Fatalf("error message %q contains stdin token", err.Error())
	}
}

func TestLaunch_usesDefaultCommandRunner_whenRunIsNil(t *testing.T) {
	// Given: PATH 上の fake agentsecrets と dummy project。run は nil（実 exec）
	projectDir := setupFakeAgentsecretsOnPATH(t)
	launcher := agentsecrets.NewLauncher(
		projectDir,
		[]string{secretnames.CursorAPIKeyName},
		[]string{"PATH", "HOME", "TMPDIR"},
		nil,
		nil,
	)

	// When: stdin を stdout へ流す孫を起動する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "cat",
		Stdin:   []byte("default-runner-body"),
	})

	// Then: 実 exec 経路で stdout が返る
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	if string(got) != "default-runner-body" {
		t.Fatalf("Launch() = %q, want default-runner-body", string(got))
	}
}

func TestLaunch_defaultCommandRunnerOmitsStderrFromError_whenChildExitsNonZero(t *testing.T) {
	// Given: run=nil と、stderr へ書いて失敗する孫
	projectDir := setupFakeAgentsecretsOnPATH(t)
	const stderrToken = "agentsecrets-default-runner-stderr-token"
	launcher := agentsecrets.NewLauncher(projectDir, nil, []string{"PATH", "HOME", "TMPDIR"}, nil, nil)

	// When: 失敗孫を起動する
	_, err := launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "sh",
		Args:    []string{"-c", "printf %s " + stderrToken + " 1>&2; exit 1"},
	})

	// Then: error に stderr 本文は出ない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if strings.Contains(err.Error(), stderrToken) {
		t.Fatalf("error message %q contains stderr token", err.Error())
	}
}

func TestLaunch_failsWhenLauncherIsNil(t *testing.T) {
	t.Parallel()

	// Given: nil receiver
	var launcher *agentsecrets.Launcher

	// When: Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "agent"})

	// Then: Infrastructure Error
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
	var infra *agentsecrets.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *agentsecrets.Error", err, err)
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestLaunch_failsWhenCtxIsNil(t *testing.T) {
	t.Parallel()

	// Given: 妥当な launcher
	projectDir := filepath.Join(t.TempDir(), "cursor-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	launcher := agentsecrets.NewLauncher(projectDir, nil, nil, nil, func(
		context.Context, string, []string, string, []string, []byte,
	) ([]byte, error) {
		return nil, errors.New("runner must not start")
	})

	// When: nil ctx
	got, err := launcher.Launch(nil, commandlaunch.Command{Program: "agent"})

	// Then: 起動前失敗
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
}
