package agentsecrets

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

// CommandRunner は wrapper 起動の差し替え可能な exec 境界である。
// name / args / dir / env / stdin は秘密値を含まない。戻り error に stdin・stderr 本文を載せない契約は呼び出し側が守る。
type CommandRunner func(
	ctx context.Context,
	name string,
	args []string,
	dir string,
	env []string,
	stdin []byte,
) ([]byte, error)

// Launcher は AgentSecrets EnvWrapper 経由で command を起動する。
// Go process は秘密値を解決せず、注入は wrapper の ProjectDir が指す project に閉じる。
type Launcher struct {
	wrapper               EnvWrapper
	inheritedEnvNameAllow []string
	lookupEnv             func(key string) (string, bool)
	run                   CommandRunner
}

// NewLauncher は AgentSecrets 実装の Launcher を返す。
//
// @require projectDir は絶対 path（Validate で起動前に弾く）。inheritedEnvNameAllow は Composition 所有の名前集合。
// @ensure 戻りは commandlaunch.Launcher。lookupEnv が nil のとき os.LookupEnv、run が nil のとき実 exec を使う。
// @ensure SecretKeys は argv / child env へ載せない（宣言のみ）。
func NewLauncher(
	projectDir string,
	secretKeys []string,
	inheritedEnvNameAllow []string,
	lookupEnv func(key string) (string, bool),
	run CommandRunner,
) *Launcher {
	keys := make([]string, len(secretKeys))
	copy(keys, secretKeys)
	allow := make([]string, len(inheritedEnvNameAllow))
	copy(allow, inheritedEnvNameAllow)
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	if run == nil {
		run = defaultCommandRunner
	}
	return &Launcher{
		wrapper: EnvWrapper{
			ProjectDir: projectDir,
			SecretKeys: keys,
		},
		inheritedEnvNameAllow: allow,
		lookupEnv:             lookupEnv,
		run:                   run,
	}
}

// Launch は command を agentsecrets env -- で包んで起動し、成功時は stdout bytes を返す。
//
// @require command.Program は trim 後に非空。wrapper.ProjectDir は Validate を満たす。
// @ensure 失敗時の error に秘密値・stdin・child stderr 本文を含めない。
// @ensure child（wrapper）env は allowlist で親から拾った entry だけであり、親環境を全継承しない。
// @ensure Go process は secret 値を env へ載せない。
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
	if err := l.wrapper.Validate(); err != nil {
		return nil, infraErr("validate_project_dir", err)
	}

	name, args := l.wrapper.Command(append([]string{program}, command.Args...))
	env := buildInheritedEnv(l.inheritedEnvNameAllow, l.lookupEnv)

	stdout, err := l.run(ctx, name, args, l.wrapper.ProjectDir, env, command.Stdin)
	if err != nil {
		return nil, infraErr("run", err)
	}
	return stdout, nil
}

func buildInheritedEnv(allow []string, lookup func(key string) (string, bool)) []string {
	env := make([]string, 0, len(allow))
	for _, name := range allow {
		value, ok := lookup(name)
		if !ok {
			continue
		}
		env = append(env, name+"="+value)
	}
	return env
}

func defaultCommandRunner(
	ctx context.Context,
	name string,
	args []string,
	dir string,
	env []string,
	stdin []byte,
) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	// why: nil Env は親環境の全継承を意味する。空でも非 nil を渡して継承を断つ。
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	// why: stderr 本文を読まず error へ写さない契約を、Discard で明示する。
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}
