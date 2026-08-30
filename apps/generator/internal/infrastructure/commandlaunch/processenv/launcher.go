// Package processenv は注入された secret と親環境から、child process の environment を組む。
package processenv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
)

var _ commandlaunch.Launcher = (*Launcher)(nil)

// Launcher は Composition が渡した secret と親 environ から child environment を組み立てる。
type Launcher struct {
	secret    commandlaunch.SecretEnv
	lookupEnv func(key string) (string, bool)
	environ   func() []string
}

// NewLauncher は process-env 実装の Launcher を返す。
//
// @require secret.Name は非空。secret.Value は検証済みの秘密値。lookupEnv は非 nil（nil のとき Launch が error を返す）。
// @ensure 戻りは commandlaunch.Launcher。
func NewLauncher(
	secret commandlaunch.SecretEnv,
	lookupEnv func(key string) (string, bool),
) *Launcher {
	return &Launcher{
		secret:    secret,
		lookupEnv: lookupEnv,
		environ:   os.Environ,
	}
}

// NewSecretEnvLauncherFactory は secret 値と親環境アクセス手段を閉じ込め、
// inject env 名を受け取って Launcher を組む factory を返す。
//
// @require secretValue は検証済みの秘密値。lookupEnv は非 nil（nil のとき Launch が error を返す）。
// @ensure 戻りは commandlaunch.SecretEnvLauncherFactory。secret 値は closure に閉じ、戻り値の呼び出し側へ渡らない。
func NewSecretEnvLauncherFactory(
	secretValue string,
	lookupEnv func(key string) (string, bool),
) commandlaunch.SecretEnvLauncherFactory {
	return func(envName string) commandlaunch.Launcher {
		return NewLauncher(
			commandlaunch.SecretEnv{Name: envName, Value: secretValue},
			lookupEnv,
		)
	}
}

// Launch は command を起動し、成功時は stdout bytes を返す。
//
// @require command.Program は trim 後に非空。lookupEnv が注入済み（nil なら error を返す）。
// @ensure 失敗時の error に秘密値・stdin・child stderr 本文を含めない。
// @ensure child env は親 environ を継承し、inject secret で同名を上書きしたものである。
func (l *Launcher) Launch(ctx context.Context, command commandlaunch.Command) ([]byte, error) {
	if l == nil {
		return nil, infraErr("launch", fmt.Errorf("launcher is nil"))
	}
	if ctx == nil {
		return nil, infraErr("launch", fmt.Errorf("ctx is nil"))
	}
	if l.lookupEnv == nil {
		return nil, infraErr("launch", fmt.Errorf("lookupEnv is nil"))
	}
	program := strings.TrimSpace(command.Program)
	if program == "" {
		return nil, infraErr("launch", fmt.Errorf("program is empty"))
	}

	parent := []string(nil)
	if l.environ != nil {
		parent = l.environ()
	}
	env := buildChildEnv(parent, l.secret.Name, l.secret.Value)

	cmd := exec.CommandContext(ctx, program, command.Args...)
	// why: nil Env は親環境の全継承を意味する。空でも非 nil を渡して継承を断つ。
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(command.Stdin)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	// why: stderr 本文を読まず error へ写さない契約を、未読 buffer ではなく Discard で明示する。
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, infraErr("run", err)
	}
	return stdout.Bytes(), nil
}

// buildChildEnv は親 environ を写し、inject 名を除いたうえで secret を載せる。
// why: PR80 probe は親 env 継承で Cursor API 到達に成功し、allowlist のみ（env -i 相当）では
//
//	「Failed to reach the Cursor API」になる。Composition の結線形は変えない。
func buildChildEnv(parent []string, secretName string, secretValue string) []string {
	env := make([]string, 0, len(parent)+1)
	for _, entry := range parent {
		name, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if name == secretName {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, secretName+"="+secretValue)
	return env
}
