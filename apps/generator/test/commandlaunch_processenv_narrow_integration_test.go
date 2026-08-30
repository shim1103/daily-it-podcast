// Scope: Narrow Integration
// 実物境界: processenv.Launcher が起動する child process（test 用 script / env）
// Double: secret は検証済みの秘密値を直接渡す。本番 credential は使わない。
// @require Launcher に注入済みの秘密値と denied parent env を契約で検証する。child は controllable な script。
// @ensure child env は親 environ を継承し、denied 名と親だけの secret は載せない。inject secret は上書きされる。
// @ensure error message に secret 値・stdin・child stderr 本文が出ない。
// @invariant 他 vendor の secret は child environ へ継承されない。
package test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/config"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
)

const (
	narrowSecretName       = "PROCESSENV_NARROW_TEST_SECRET_KEY"
	narrowDummySecretValue = "processenv-narrow-integration-dummy-secret-value"
	narrowHarmlessEnv      = "PROCESSENV_NARROW_HARMLESS"
	narrowHarmlessValue    = "narrow-harmless-passthrough"
	narrowDeniedValue      = "narrow-denied-secret-token"
	narrowStdinToken       = "narrow-integration-stdin-token"
	narrowStderrToken      = "narrow-integration-stderr-token"
)

func newNarrowLauncher(t *testing.T) *processenv.Launcher {
	t.Helper()
	return processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: narrowSecretName, Value: narrowDummySecretValue},
		os.LookupEnv,
		[]string{
			config.GetXAPIKeyEnv,
			config.GeminiAPIKeyEnv,
			config.GoogleOAuthClientSecretEnv,
			config.GoogleOAuthRefreshTokenEnv,
		},
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

func TestProcessEnvLauncher_inheritsParentExceptDeniedSecrets_whenChildPrintsEnviron(t *testing.T) {
	// Given: 無害な親 env・denied vendor secret・注入 secret
	t.Setenv(narrowHarmlessEnv, narrowHarmlessValue)
	t.Setenv(config.GetXAPIKeyEnv, narrowDeniedValue)
	launcher := newNarrowLauncher(t)

	// When: environ を stdout へ出す実 child を起動する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "env"})

	// Then: 無害 env と inject secret は渡り、denied vendor secret は無い
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	out := string(got)
	if !strings.Contains(out, narrowSecretName+"="+narrowDummySecretValue) {
		t.Fatalf("child environ = %q, want injected secret entry", out)
	}
	if !strings.Contains(out, narrowHarmlessEnv+"="+narrowHarmlessValue) {
		t.Fatalf("child environ = %q, want harmless parent entry", out)
	}
	if strings.Contains(out, config.GetXAPIKeyEnv+"=") || strings.Contains(out, narrowDeniedValue) {
		t.Fatalf("child environ = %q, want no denied vendor secret", out)
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
