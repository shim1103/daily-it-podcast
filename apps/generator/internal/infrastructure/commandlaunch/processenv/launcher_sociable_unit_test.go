package processenv_test

import (
	"context"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch/processenv"
)

const (
	testSecretName  = "PROCESSENV_TEST_SECRET_KEY"
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

	// Then: 非 nil の Launcher が返る
	if launcher == nil {
		t.Fatal("factory() returned nil, want Launcher")
	}
}
