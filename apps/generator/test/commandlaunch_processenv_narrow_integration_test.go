// Scope: Narrow Integration
// 実物境界: processenv.Launcher が起動する child process（test 用 script / env）
// Double: secret は検証済みの秘密値を直接渡す。本番 credential は使わない。
// allowlist SSoT: commandlaunch.InheritedEnvNameAllow()（contract package アクセッサ）
// @require Launcher に注入済みの秘密値と allowlist を契約で検証する。child は controllable な script。
// @ensure child env は commandlaunch.InheritedEnvNameAllow() + 秘密値だけ。未設定 secret / 空 program は起動前失敗。
// @ensure error message に secret 値・stdin・child stderr 本文が出ない。
// @invariant 親固有の secret は child environ へ継承されない。
package test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/manuscript/cursorcli"
)

const (
	narrowDummySecretValue = "narrow-integration-dummy-cursor-api-key"
	narrowParentOnlySecret = "NARROW_PARENT_ONLY_SECRET"
	narrowParentOnlyValue  = "narrow-parent-only-secret-token"
	narrowStdinToken       = "narrow-integration-stdin-token"
	narrowStderrToken      = "narrow-integration-stderr-token"
)

func newNarrowLauncher(t *testing.T) *processenv.Launcher {
	t.Helper()
	// production runtime（os.LookupEnv）を injection して Launcher を構築
	return processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: cursorcli.CursorAPIKeyEnvName, Value: narrowDummySecretValue},
		os.LookupEnv,
	)
}

func writeNarrowMarkerChild(t *testing.T) (program string, marker string) {
	t.Helper()
	dir := t.TempDir()
	marker = filepath.Join(dir, "started")
	program = filepath.Join(dir, "child.sh")
	script := "#!/bin/sh\ntouch '" + marker + "'\nprintf 'ok'\n"
	if err := os.WriteFile(program, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v, want nil", err)
	}
	return program, marker
}

func TestProcessEnvLauncher_passesOnlyAllowlistAndCursorSecret_whenChildPrintsEnviron(t *testing.T) {
	// Given: contract allowlist SSoT と Cursor secret、および親固有 secret
	t.Setenv(narrowParentOnlySecret, narrowParentOnlyValue)
	launcher := newNarrowLauncher(t)

	// When: environ を stdout へ出す実 child を起動する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "env"})

	// Then: commandlaunch.InheritedEnvNameAllow + Cursor secret だけが child に渡り、親固有 secret は無い
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	out := string(got)
	if !strings.Contains(out, cursorcli.CursorAPIKeyEnvName+"="+narrowDummySecretValue) {
		t.Fatalf("child environ = %q, want Cursor secret entry", out)
	}
	if strings.Contains(out, narrowParentOnlySecret) || strings.Contains(out, narrowParentOnlyValue) {
		t.Fatalf("child environ = %q, want no parent-only secret", out)
	}
	for _, name := range commandlaunch.InheritedEnvNameAllow() {
		if value, ok := os.LookupEnv(name); ok {
			entry := name + "=" + value
			if !strings.Contains(out, entry) {
				t.Fatalf("child environ = %q, want allowlist entry %q", out, entry)
			}
		}
	}
}

func TestProcessEnvLauncher_failsBeforeChildStart_whenProgramIsEmpty(t *testing.T) {
	// Given: secret は注入済み
	launcher := newNarrowLauncher(t)
	_, marker := writeNarrowMarkerChild(t)

	// When: 空 Program で Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: ""})

	// Then: 起動前失敗。child は走らない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("marker %q exists, want child not started", marker)
	}

	// then: error は *processenv.Error で契約を満たす
	var infraErr *processenv.Error
	if !errors.As(err, &infraErr) {
		t.Fatalf("errors.As(err, &infraErr) = false, want true; type = %T", err)
	}
	errMsg := infraErr.Error()
	if !strings.HasPrefix(errMsg, "processenv:") {
		t.Fatalf("error message = %q, want \"processenv:\" prefix", errMsg)
	}
	if infraErr.Unwrap() == nil {
		t.Fatal("infraErr.Unwrap() = nil, want non-nil")
	}
}

func TestProcessEnvLauncher_errorOmitsSecretStdinAndStderr_whenChildExitsNonZero(t *testing.T) {
	// Given: 識別可能な secret・stdin・stderr を持つ失敗 child
	launcher := newNarrowLauncher(t)

	// When: stderr へ書いて非0 exit する実 child を起動する
	_, err := launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "sh",
		Args:    []string{"-c", "printf %s " + narrowStderrToken + " 1>&2; exit 1"},
		Stdin:   []byte(narrowStdinToken),
	})

	// Then: error は返るが secret・stdin・stderr 本文は message に出ない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	msg := err.Error()
	for _, leaked := range []string{narrowDummySecretValue, narrowStdinToken, narrowStderrToken} {
		if strings.Contains(msg, leaked) {
			t.Fatalf("error message %q contains leaked token %q", msg, leaked)
		}
	}
}
