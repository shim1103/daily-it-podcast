package processenv_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

const (
	testSecretName  = "PROCESSENV_TEST_CURSOR_API_KEY"
	testSecretValue = "processenv-test-dummy-secret-value"
	testAllowPATH   = "PROCESSENV_TEST_ALLOW_PATH"
	testAllowHOME   = "PROCESSENV_TEST_ALLOW_HOME"
)

type testBindings map[secrettransport.SecretRef]string

func (b testBindings) ResolveSecret(ref secrettransport.SecretRef) (string, bool) {
	name, ok := b[ref]
	return name, ok
}

func newTestLauncher(t *testing.T, allow ...string) *processenv.Launcher {
	t.Helper()
	ref := secrettransport.NewSecretRef()
	bindings := testBindings{ref: testSecretName}
	if len(allow) == 0 {
		allow = []string{testAllowPATH, testAllowHOME}
	}
	return processenv.NewLauncher(bindings, ref, allow, nil)
}

func writeMarkerChild(t *testing.T) (program string, marker string) {
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

func TestLaunch_returnsStdout_whenChildSucceeds(t *testing.T) {
	// Given: secret が設定された launcher
	t.Setenv(testSecretName, testSecretValue)
	t.Setenv(testAllowPATH, "/processenv-test-bin")
	launcher := newTestLauncher(t)

	// When: stdin を stdout へ流す決定的な command を起動する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{
		Program: "cat",
		Stdin:   []byte("envelope-body"),
	})

	// Then: stdout が返り、error は無い
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	if string(got) != "envelope-body" {
		t.Fatalf("Launch() = %q, want stdin echoed to stdout", string(got))
	}
}

func TestLaunch_failsBeforeChildStart_whenSecretValueIsEmpty(t *testing.T) {
	// Given: secret 名は存在するが値が空
	t.Setenv(testSecretName, "")
	t.Setenv(testAllowPATH, "/processenv-test-bin")
	launcher := newTestLauncher(t)
	program, marker := writeMarkerChild(t)

	// When: Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: program})

	// Then: 起動前に失敗し、child は走らない。error は Infrastructure Error として判別できる
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil on failure", string(got))
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("marker %q exists, want child not started", marker)
	}
	var infra *processenv.Error
	if !errors.As(err, &infra) {
		t.Fatalf("error type %T (%v), want *processenv.Error", err, err)
	}
	if !strings.HasPrefix(infra.Error(), "processenv:") {
		t.Fatalf("Error() = %q, want prefix processenv:", infra.Error())
	}
	if errors.Unwrap(infra) == nil {
		t.Fatal("Unwrap() is nil")
	}
}

func TestLaunch_failsBeforeChildStart_whenSecretBindingIsUnresolved(t *testing.T) {
	// Given: BindingResolver が解決できない SecretRef
	t.Setenv(testSecretName, testSecretValue)
	unresolved := secrettransport.NewSecretRef()
	launcher := processenv.NewLauncher(testBindings{}, unresolved, []string{testAllowPATH}, nil)
	program, marker := writeMarkerChild(t)

	// When: Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: program})

	// Then: 起動前に失敗し、child は走らない
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil on failure", string(got))
	}
	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("marker %q exists, want child not started", marker)
	}
}

func TestLaunch_passesClosureResolvedSecretValue_whenLookupEnvInjectedDirectlyAsClosure(t *testing.T) {
	t.Parallel()

	// Given: t.Setenv を使わず、closure で直接注入した lookupEnv が解決する secret
	ref := secrettransport.NewSecretRef()
	const secretName = "PROCESSENV_TEST_CLOSURE_SECRET_KEY"
	const secretValue = "closure-secret-real-value"
	bindings := testBindings{ref: secretName}
	lookupEnv := func(key string) (string, bool) {
		if key == secretName {
			return secretValue, true
		}
		return "", false
	}
	launcher := processenv.NewLauncher(bindings, ref, nil, lookupEnv)

	// When: environ を stdout へ出す実 child を起動する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "env"})

	// Then: closure が返した実値が secret として child environ に渡る
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	out := string(got)
	if !strings.Contains(out, secretName+"="+secretValue) {
		t.Fatalf("child environ = %q, want closure-resolved secret entry", out)
	}
}
