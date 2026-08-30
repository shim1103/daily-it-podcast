// Package processenv は注入された secret と親環境アクセス手段で child environment を組み、command を起動する。
package processenv

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
)

var _ commandlaunch.Launcher = (*Launcher)(nil)

// Launcher は Composition が渡した secret と contract allowlist だけで child environment を組み立てる。
type Launcher struct {
	secret    commandlaunch.SecretEnv
	lookupEnv func(key string) (string, bool)
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
// @ensure child env は commandlaunch.InheritedEnvNameAllow で親から拾った entry と秘密値だけであり、親環境を全継承しない。
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

	env := buildChildEnv(commandlaunch.InheritedEnvNameAllow(), l.secret.Name, l.secret.Value, l.lookupEnv)

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

func buildChildEnv(
	allow []string,
	secretName string,
	secretValue string,
	lookup func(key string) (string, bool),
) []string {
	env := make([]string, 0, len(allow)+1)
	for _, name := range allow {
		value, ok := lookup(name)
		if !ok {
			continue
		}
		env = append(env, name+"="+value)
	}
	env = append(env, secretName+"="+secretValue)
	return env
}
