package cursorcli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/application/port"
	"github.com/shim1103/daily-it-podcast/apps/generator/internal/infrastructure/commandlaunch"
)

var _ port.TextWriter = (*TextWriter)(nil)

// TextWriter は Cursor CLI を commandlaunch.Launcher 経由で実行する Adapter である。
type TextWriter struct {
	launcher commandlaunch.Launcher
}

// NewTextWriter は Launcher factory から TextWriter を組み立てる。
// cursorcli は自身が所有する inject env 名（N2、CursorAPIKeyEnvName）を factory へ渡し、
// secret 値・親環境アクセス手段・runtime 実装を知らない。
//
// @require newLauncher は非 nil。
// @ensure 戻りは port.TextWriter。cursorcli は secret 値を保持しない。Command には Program・Args・Stdin のみ載る。
func NewTextWriter(newLauncher commandlaunch.SecretEnvLauncherFactory) *TextWriter {
	return &TextWriter{launcher: newLauncher(CursorAPIKeyEnvName)}
}

// Write は port.TextWriter の実装。契約の基本は port.TextWriter を参照する。
// 以下は Cursor CLI 起動という実行手段に固有の追加契約だけを示す。
//
// @require launcher が非 nil。brief は trim 後に非空。
// @ensure 失敗時は必ず *cursorcli.Error を返し、断片は空文字。
// @ensure Launch へ渡す Command は Program・Args・Stdin のみで、秘密・project・runtime を含まない。
// @ensure ctx の cancel は launcher へ伝播する。
// @invariant stderr の内容文字列を戻り error へ写さない。ManuscriptDraft へ変換しない（Application の責務）。
func (w *TextWriter) Write(ctx context.Context, brief string) (string, error) {
	if w == nil || w.launcher == nil {
		return "", infraErr("write", fmt.Errorf("launcher is nil"))
	}

	trimmed := strings.TrimSpace(brief)
	if trimmed == "" {
		return "", infraErr("validate_brief", fmt.Errorf("brief is empty after trim"))
	}

	stdout, err := w.launcher.Launch(ctx, commandlaunch.Command{
		Program: BinaryName,
		Args:    buildCursorArgs(),
		Stdin:   []byte(trimmed),
	})
	if err != nil {
		return "", infraErr("run", err)
	}

	fragment, err := decodeFragment(stdout)
	if err != nil {
		return "", infraErr("decode_envelope", err)
	}
	return fragment, nil
}

// buildCursorArgs は constants.go の確定値だけから Cursor CLI の argv を決定的に構築する。
// 秘密値は argv に載せない。
func buildCursorArgs() []string {
	return []string{
		PrintFlag,
		ModeFlag, Mode,
		OutputFlag, OutputFormat,
		ModelFlag, ModelID,
		SandboxFlag, SandboxValue,
		TrustFlag,
	}
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
