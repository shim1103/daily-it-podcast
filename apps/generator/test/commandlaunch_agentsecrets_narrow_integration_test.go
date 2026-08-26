// Scope: Narrow Integration
// 実物境界: agentsecrets.Launcher が起動する child process（controllable fake `agentsecrets`）
// Double: 本番 keychain / 実 agentsecrets binary は使わない。Composition allowlist / project SSoT を読む。
// @require dummy project dir と PATH 上の fake agentsecrets を用意する。
// @ensure wrapper cwd は Cursor 専用 project。child env は allowlist + fake 注入 secret のみ。
// @ensure error message に stdin・child stderr 本文が出ない。親固有 secret は wrapper environ へ継承されない。
package test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/composition"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

const (
	asNarrowDummySecretValue = "as-narrow-dummy-cursor-api-key"
	asNarrowParentOnlySecret = "AS_NARROW_PARENT_ONLY_SECRET"
	asNarrowParentOnlyValue  = "as-narrow-parent-only-secret-token"
	asNarrowStdinToken       = "as-narrow-integration-stdin-token"
	asNarrowStderrToken      = "as-narrow-integration-stderr-token"
)

type agentsecretsNarrowFixture struct {
	projectDir string
	launcher   *agentsecrets.Launcher
}

// setupAgentsecretsNarrowFixture は dummy HOME / Cursor project / fake agentsecrets PATH を用意する。
func setupAgentsecretsNarrowFixture(t *testing.T) agentsecretsNarrowFixture {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	projectDir := composition.CursorCommandProjectDir()
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v, want nil", err)
	}
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(bin) error = %v, want nil", err)
	}
	writeFakeAgentsecretsEnv(t, binDir)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return agentsecretsNarrowFixture{
		projectDir: projectDir,
		launcher: agentsecrets.NewLauncher(
			projectDir,
			[]string{secretnames.CursorAPIKeyName},
			composition.CursorCommandInheritedEnvNameAllow(),
			nil,
			nil,
		),
	}
}

func writeFakeAgentsecretsEnv(t *testing.T, binDir string) {
	t.Helper()
	// what: `agentsecrets env -- <cmd>...` を模擬する。keychain は読まず、固定 dummy secret を孫へ注入する。
	script := `#!/bin/sh
set -eu
if [ "$1" != "env" ] || [ "$2" != "--" ]; then
  printf 'fake-agentsecrets: unexpected argv\n' >&2
  exit 2
fi
shift 2
# why: allowlist だけを親から受け取り、secret は wrapper が注入する契約を模擬する。
export ` + secretnames.CursorAPIKeyName + `='` + asNarrowDummySecretValue + `'
exec "$@"
`
	path := filepath.Join(binDir, agentsecrets.EnvBinary)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v, want nil", path, err)
	}
}

func TestAgentsecretsLauncher_runsChildInCursorProjectDir_whenFakeEnvWrapperIsOnPath(t *testing.T) {
	// Given: dummy Cursor project と PATH 上の fake agentsecrets
	fx := setupAgentsecretsNarrowFixture(t)
	t.Setenv(asNarrowParentOnlySecret, asNarrowParentOnlyValue)

	// When: cwd を stdout へ出す孫 command を Launch する
	got, err := fx.launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "pwd",
	})

	// Then: wrapper の cwd（= project dir）で孫が動く
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	gotPwd := strings.TrimSpace(string(got))
	wantPwd, evalErr := filepath.EvalSymlinks(fx.projectDir)
	if evalErr != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) error = %v, want nil", fx.projectDir, evalErr)
	}
	gotResolved, evalErr := filepath.EvalSymlinks(gotPwd)
	if evalErr != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) error = %v, want nil", gotPwd, evalErr)
	}
	if gotResolved != wantPwd {
		t.Fatalf("child pwd = %q (resolved %q), want projectDir %q (resolved %q)", gotPwd, gotResolved, fx.projectDir, wantPwd)
	}
}

func TestAgentsecretsLauncher_passesAllowlistAndInjectedSecretOnly_whenChildPrintsEnviron(t *testing.T) {
	// Given: Composition allowlist SSoT・dummy project・親固有 secret
	fx := setupAgentsecretsNarrowFixture(t)
	t.Setenv(asNarrowParentOnlySecret, asNarrowParentOnlyValue)
	allow := composition.CursorCommandInheritedEnvNameAllow()

	// When: environ を stdout へ出す孫を起動する
	got, err := fx.launcher.Launch(context.Background(), commandlaunch.Command{Program: "env"})

	// Then: allowlist + fake 注入 Cursor secret だけ。親固有 secret は無い
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	out := string(got)
	if !strings.Contains(out, secretnames.CursorAPIKeyName+"="+asNarrowDummySecretValue) {
		t.Fatalf("child environ = %q, want Cursor secret entry from fake wrapper", out)
	}
	if strings.Contains(out, asNarrowParentOnlySecret) || strings.Contains(out, asNarrowParentOnlyValue) {
		t.Fatalf("child environ = %q, want no parent-only secret", out)
	}
	for _, name := range allow {
		if value, ok := os.LookupEnv(name); ok {
			entry := name + "=" + value
			if !strings.Contains(out, entry) {
				t.Fatalf("child environ = %q, want allowlist entry %q", out, entry)
			}
		}
	}
}

func TestAgentsecretsLauncher_failsBeforeChildStart_whenProgramIsEmpty(t *testing.T) {
	// Given: dummy project と、起動すると marker を書く孫（到達してはならない）
	fx := setupAgentsecretsNarrowFixture(t)
	marker := filepath.Join(t.TempDir(), "started")
	program := filepath.Join(t.TempDir(), "child.sh")
	script := "#!/bin/sh\ntouch '" + marker + "'\n"
	if err := os.WriteFile(program, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v, want nil", err)
	}

	// When: 空 Program で Launch する
	got, err := fx.launcher.Launch(context.Background(), commandlaunch.Command{Program: ""})

	// Then: 起動前失敗。marker は無い
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("marker %q exists, want child not started", marker)
	}
}

func TestAgentsecretsLauncher_errorOmitsStdinAndStderr_whenChildExitsNonZero(t *testing.T) {
	// Given: 識別可能な stdin / stderr を持つ失敗孫
	fx := setupAgentsecretsNarrowFixture(t)

	// When: stderr へ書いて非0 exit する孫を起動する
	_, err := fx.launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "sh",
		Args:    []string{"-c", "printf %s " + asNarrowStderrToken + " 1>&2; exit 1"},
		Stdin:   []byte(asNarrowStdinToken),
	})

	// Then: error は返るが stdin・stderr 本文・dummy secret 値は message に出ない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	msg := err.Error()
	for _, leaked := range []string{asNarrowDummySecretValue, asNarrowStdinToken, asNarrowStderrToken} {
		if strings.Contains(msg, leaked) {
			t.Fatalf("error message %q contains leaked token %q", msg, leaked)
		}
	}
}
