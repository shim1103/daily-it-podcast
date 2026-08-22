package cursorcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/agentsecrets"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/secretnames"
)

var _ port.TextWriter = (*TextWriter)(nil)

// runFunc は Cursor CLI を 1 度実行し、stdout を返す。
// dir は子 process の working directory で、秘密供給 wrapper の active project を決める。
// env は子 process へ渡す環境変数そのもので、親環境を継承しない。
// stderr は内部診断として扱い、戻り値へ載せない。
type runFunc func(ctx context.Context, name string, args []string, dir string, env []string, stdin string) ([]byte, error)

type TextWriter struct {
	// wrapper は CLI 境界の秘密供給。Cursor へ渡る秘密の範囲もここが決める。
	wrapper agentsecrets.EnvWrapper
	runFn   runFunc // why: test の並列実行と共存するため package global に置かない
}

// projectName は Cursor 専用 AgentSecrets project の名前。
//
// why: どの project を使うかは「Cursor を呼ぶ」という Adapter の責務に属する。dir の
// 配置規約と home の解決は agentsecrets が所有するため、ここは名前だけを持つ。
const projectName = "cursor"

// NewTextWriter は Cursor CLI を実行する TextWriter Adapter を返す。
//
// why: 既存 adapter が `&agentsecrets.Client{}` をゼロ値で受け取り、接続先の解決を
// agentsecrets 側（DefaultProxyURL）へ委ねているのと同型にする。呼び出し側へ位置情報の
// 組み立てを要求しない。
//
// @ensure 秘密値は保持しない。依存する秘密は名前でだけ宣言する。
// @ensure Cursor 専用 project の設定 dir を wrapper の working directory にする。
func NewTextWriter() *TextWriter {
	return &TextWriter{
		wrapper: agentsecrets.EnvWrapper{
			ProjectDir: agentsecrets.DefaultProjectDir(projectName),
			// why: 既存 5 adapter が Inject へ key 名を渡して依存を明示するのと同型に、
			// CLI 境界でも依存する秘密を code へ残す。実注入は project 分離が担うため、
			// この宣言は実行時の入力にはならない（EnvWrapper.SecretKeys の warn を参照）。
			SecretKeys: []string{secretnames.CursorAPIKeyName},
		},
		runFn: runCursorCLI,
	}
}

// RequiredSecretKeys はこの Adapter が依存する秘密キー名を返す。
//
// warn: ⚠️ 戻りは宣言であって、実際に子 process へ注入される秘密の一覧ではない。
// `agentsecrets env` は key 名で絞れないため、実注入は wrapper の ProjectDir が指す
// project の全 secret になる。両者が一致するかは運用側の project 構成が決める。
//
// @ensure 戻りの要素に秘密値を含まない。名前だけを返す。
func (w *TextWriter) RequiredSecretKeys() []string {
	if w == nil {
		return nil
	}
	return slices.Clone(w.wrapper.SecretKeys)
}

// Write は port.TextWriter の実装。契約の基本は port.TextWriter を参照する。
// 以下は Cursor CLI 起動という実行手段に固有の追加契約だけを示す。
//
// @require agentsecrets.EnvBinary と BinaryName の実行ファイルが PATH にある。PATH 不在は実行時 error として畳む。
// @require 組み立て時に受け取った project dir が存在し、Cursor 専用 AgentSecrets project の設定 file を持つ。
// @require Cursor の秘密は agentsecrets.EnvBinary が keychain から解決する。Go process は秘密値を保持しない。
// @ensure 失敗時は必ず *cursorcli.Error を返し、断片は空文字列。
// @ensure 子 process の working directory は project dir。渡る秘密の範囲はその project の secret に限られる。
// @ensure ctx の cancel は起動済み process へ伝播する。
// @invariant stderr の内容文字列を戻り error へ写さない（decision §5）。
func (w *TextWriter) Write(ctx context.Context, brief string) (string, error) {
	if w == nil || w.runFn == nil {
		return "", infraErr("write", fmt.Errorf("runner is nil"))
	}

	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return "", infraErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	if err := w.wrapper.Validate(); err != nil {
		return "", infraErr("validate_secret_boundary", err)
	}

	name, args := w.wrapper.Command(buildCursorArgs())
	stdout, err := w.runFn(ctx, name, args, w.wrapper.ProjectDir, buildMinimalEnv(), trimmed)
	if err != nil {
		// why: decision §5。stderr の内容文字列は写さず、exit 由来の error だけを cause に残す。
		return "", infraErr("run", err)
	}

	fragment, err := decodeFragment(stdout)
	if err != nil {
		return "", infraErr("decode_envelope", err)
	}
	return fragment, nil
}

// buildCursorArgs は constants.go の確定値だけから Cursor CLI の argv を決定的に構築する。
// 秘密値は argv に載せない。wrapper が子 process の env へ注入する。
func buildCursorArgs() []string {
	return []string{
		BinaryName,
		PrintFlag,
		ModeFlag, Mode,
		OutputFlag, OutputFormat,
		ModelFlag, ModelID,
		SandboxFlag, SandboxValue,
		TrustFlag,
	}
}

// minimalEnvNames は子 process へ引き継ぐ環境変数の名前。値は親から取るが、
// ここに無い名前は落とす（decision 2026-08-22T11-55-22 §3）。
//
// why: PATH が無いと wrapper が BinaryName を解決できず、HOME が無いと wrapper の
// project 紐付けと Cursor CLI の設定 dir（どちらも $HOME 配下）が読めない。
// TMPDIR は macOS の per-user 一時領域で、落とすと子 process の一時 file が
// 共有 /tmp へ逃げる。この 3 つが Cursor 呼び出しの正当な目的に必要な最小で、
// 他 vendor の秘密を含む親環境の残りは渡さない。
// why not: TERM は非対話の PrintFlag 実行に不要なため含めない。
var minimalEnvNames = []string{"PATH", "HOME", "TMPDIR"}

// buildMinimalEnv は minimalEnvNames の名前だけを親環境から拾って env を構築する。
// 親が定義していない名前は空値で載せず、entry ごと落とす。
func buildMinimalEnv() []string {
	env := make([]string, 0, len(minimalEnvNames))
	for _, name := range minimalEnvNames {
		value, ok := os.LookupEnv(name)
		if !ok {
			continue
		}
		env = append(env, name+"="+value)
	}
	return env
}

func runCursorCLI(ctx context.Context, name string, args []string, dir string, env []string, stdin string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	// why: AgentSecrets の active project は cwd 直下の設定 file で決まる。ここを Cursor 専用
	// project の dir へ向けることが、子 process へ渡る秘密を絞る唯一の手段になる。
	// why not: dir の存在を事前検査しない。既存 5 adapter は全て HTTP で、proxy 未起動も
	// 接続 error として実行時に畳まれる。exec.LookPath 相当の事前検査は前例が無く、
	// 一貫性（philosophy §5-1）に従い起動失敗として同じ経路へ落とす。
	cmd.Dir = dir
	// why: exec は cmd.Env が nil のときだけ親 process の環境を全継承する。
	// buildMinimalEnv は空でも非 nil を返すため、代入するだけで継承を断てる。
	cmd.Env = env
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// why: decision §5。stderr は内部診断のため error へ写さず、exit 由来の error だけを返す。
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

func decodeFragment(stdout []byte) (string, error) {
	var parsed resultEnvelope
	if err := json.Unmarshal(stdout, &parsed); err != nil {
		return "", err
	}
	fragment := strings.TrimSpace(parsed.Result)
	if fragment == "" {
		return "", fmt.Errorf("envelope result is missing")
	}
	return fragment, nil
}

type resultEnvelope struct {
	Result string `json:"result"`
}
