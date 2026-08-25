// Package processenv は process environment から秘密を解決して command を起動する。
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
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secrettransport"
)

var _ commandlaunch.Launcher = (*Launcher)(nil)

// Launcher は Composition が渡した binding と allowlist だけで child environment を組み立てる。
type Launcher struct {
	bindings              secrettransport.BindingResolver
	secretRef             secrettransport.SecretRef
	inheritedEnvNameAllow []string
	lookupEnv             func(key string) (string, bool)
}

// NewLauncher は process-env 実装の Launcher を返す。
//
// @require bindings は非 nil。inheritedEnvNameAllow は Composition 所有の名前集合。
// @ensure 戻りは commandlaunch.Launcher。lookupEnv が nil のとき os.LookupEnv を使う。
func NewLauncher(
	bindings secrettransport.BindingResolver,
	secretRef secrettransport.SecretRef,
	inheritedEnvNameAllow []string,
	lookupEnv func(key string) (string, bool),
) *Launcher {
	allow := make([]string, len(inheritedEnvNameAllow))
	copy(allow, inheritedEnvNameAllow)
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	return &Launcher{
		bindings:              bindings,
		secretRef:             secretRef,
		inheritedEnvNameAllow: allow,
		lookupEnv:             lookupEnv,
	}
}

// Launch は command を起動し、成功時は stdout bytes を返す。
//
// @require command.Program は trim 後に非空。bindings が secretRef を解決でき、その名前の env が非空。
// @ensure 失敗時の error に秘密値・stdin・child stderr 本文を含めない。
// @ensure child env は allowlist で親から拾った entry と Cursor secret だけであり、親環境を全継承しない。
func (l *Launcher) Launch(ctx context.Context, command commandlaunch.Command) ([]byte, error) {
	if l == nil {
		return nil, infraErr("launch", fmt.Errorf("launcher is nil"))
	}
	if ctx == nil {
		return nil, infraErr("launch", fmt.Errorf("ctx is nil"))
	}
	program := strings.TrimSpace(command.Program)
	if program == "" {
		return nil, infraErr("launch", fmt.Errorf("program is empty"))
	}
	if l.bindings == nil {
		return nil, infraErr("launch", fmt.Errorf("bindings is nil"))
	}

	secretName, ok := l.bindings.ResolveSecret(l.secretRef)
	if !ok || strings.TrimSpace(secretName) == "" {
		return nil, infraErr("resolve_secret_binding", fmt.Errorf("secret binding is unresolved"))
	}
	secretValue, ok := l.lookupEnv(secretName)
	if !ok || secretValue == "" {
		return nil, infraErr("resolve_secret_value", fmt.Errorf("secret is unset"))
	}

	env := buildChildEnv(l.inheritedEnvNameAllow, secretName, secretValue, l.lookupEnv)

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
