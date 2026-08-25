package processenv_test

import (
	"context"
	"os"
	"path/filepath"
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
	return processenv.NewLauncher(bindings, ref, allow)
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

func TestLaunch_failsBeforeChildStart_whenSecretBindingIsUnresolved(t *testing.T) {
	// Given: BindingResolver が解決できない SecretRef
	t.Setenv(testSecretName, testSecretValue)
	unresolved := secrettransport.NewSecretRef()
	launcher := processenv.NewLauncher(testBindings{}, unresolved, []string{testAllowPATH})
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
