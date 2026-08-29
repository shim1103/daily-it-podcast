package processenv_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
)

const (
	testSecretName  = "PROCESSENV_TEST_CURSOR_API_KEY"
	testSecretValue = "processenv-test-dummy-secret-value"
	testAllowPATH   = "PROCESSENV_TEST_ALLOW_PATH"
	testAllowHOME   = "PROCESSENV_TEST_ALLOW_HOME"
)

func newTestLauncher(t *testing.T) *processenv.Launcher {
	t.Helper()
	// 最小限の allowlist（PATH、HOME など）を fake で供給
	lookupEnv := func(key string) (string, bool) {
		switch key {
		case testAllowPATH:
			return "/test-bin", true
		case testAllowHOME:
			return "/test-home", true
		default:
			return "", false
		}
	}
	return processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: testSecretName, Value: testSecretValue},
		lookupEnv,
	)
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

func TestLaunch_returnsNilAndError_whenLauncherIsNil(t *testing.T) {
	// Given: nil launcher
	var launcher *processenv.Launcher

	// When: Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "cat"})

	// Then: error が返り、stdout は nil
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
}

func TestLaunch_returnsNilAndError_whenContextIsNil(t *testing.T) {
	// Given: nil context
	launcher := newTestLauncher(t)

	// When: Launch する
	got, err := launcher.Launch(nil, commandlaunch.Command{Program: "cat"})

	// Then: error が返り、stdout は nil
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
}

func TestLaunch_returnsNilAndError_whenLookupEnvIsNil(t *testing.T) {
	// Given: lookupEnv が nil の launcher
	launcher := processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: testSecretName, Value: testSecretValue},
		nil,
	)

	// When: Launch する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "cat"})

	// Then: error が返り、stdout は nil（暗黙 fallback なし）
	if err == nil {
		t.Fatal("Launch() error = nil, want non-nil")
	}
	if got != nil {
		t.Fatalf("Launch() = %q, want nil", string(got))
	}
}

func TestLaunch_returnsStdout_whenChildSucceeds(t *testing.T) {
	// Given: secret が設定された launcher と、PATH resolve 用 closure
	t.Setenv(testSecretName, testSecretValue)
	t.Setenv(testAllowPATH, "/usr/bin") // 実際の PATH
	lookupEnv := func(key string) (string, bool) {
		v, ok := os.LookupEnv(key)
		return v, ok
	}
	launcher := processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: testSecretName, Value: testSecretValue},
		lookupEnv,
	)

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

func TestLaunch_passesClosureResolvedSecretValue_whenLookupEnvInjectedDirectlyAsClosure(t *testing.T) {
	t.Parallel()

	// Given: t.Setenv を使わず、closure で直接注入した lookupEnv が環境名を解決する secret
	const secretName = "PROCESSENV_TEST_CLOSURE_SECRET_KEY"
	const secretValue = "closure-secret-real-value"
	lookupEnv := func(key string) (string, bool) {
		if key == secretName {
			return secretValue, true
		}
		return "", false
	}
	launcher := processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: secretName, Value: secretValue},
		lookupEnv,
	)

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

func TestLaunch_passesContractAllowlistAndSecretOnly_whenChildPrintsEnviron(t *testing.T) {
	t.Parallel()

	// Given: contract allowlist 定義と secret
	const secretName = "PROCESSENV_TEST_CONTRACT_SECRET_KEY"
	const secretValue = "contract-secret-dummy-value"
	lookupEnv := func(key string) (string, bool) {
		switch key {
		case secretName:
			return secretValue, true
		case "PATH":
			return "/processenv-test-bin", true
		case "HOME":
			return "/processenv-test-home", true
		case "TMPDIR":
			return "/processenv-test-tmp", true
		default:
			// その他の親 env は lookup しない（allowlist にない）
			return "", false
		}
	}
	launcher := processenv.NewLauncher(
		commandlaunch.SecretEnv{Name: secretName, Value: secretValue},
		lookupEnv,
	)

	// When: environ を stdout へ出す実 child を起動する
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "env"})

	// Then: contract allowlist（PATH / HOME / TMPDIR）と secret だけが child に渡る
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	out := string(got)
	expectedEntries := []string{"PATH=/processenv-test-bin", "HOME=/processenv-test-home", "TMPDIR=/processenv-test-tmp", secretName + "=" + secretValue}
	for _, want := range expectedEntries {
		if !strings.Contains(out, want) {
			t.Fatalf("child environ = %q, want entry %q", out, want)
		}
	}
}

func TestNewSecretEnvLauncherFactory_returnsFactoryThatBuildsLauncher_withSecretAndLookupEnvClosed(t *testing.T) {
	t.Parallel()

	// Given: secret 値と lookupEnv を factory に渡す
	const secretValue = "factory-test-secret-value"
	const testEnvName = "FACTORY_TEST_ENV_NAME"
	lookupEnv := func(key string) (string, bool) {
		switch key {
		case "PATH":
			return "/factory-test-bin", true
		case "HOME":
			return "/factory-test-home", true
		default:
			return "", false
		}
	}
	factory := processenv.NewSecretEnvLauncherFactory(secretValue, lookupEnv)

	// When: factory を env 名で呼んで Launcher を得る
	launcher := factory(testEnvName)

	// Then: Launcher が返り、secret と lookupEnv が closure に閉じ込められている
	if launcher == nil {
		t.Fatal("factory() returned nil, want Launcher")
	}

	// And: Launcher を起動して env 名と secret 値が child に渡ることを確認
	got, err := launcher.Launch(context.Background(), commandlaunch.Command{Program: "env"})
	if err != nil {
		t.Fatalf("Launch() error = %v, want nil", err)
	}
	out := string(got)
	if !strings.Contains(out, testEnvName+"="+secretValue) {
		t.Fatalf("child environ = %q, want factory-closed secret entry %q=%q", out, testEnvName, secretValue)
	}
}
