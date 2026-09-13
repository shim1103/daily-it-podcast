package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// LookPath は PATH 上の実行 file を探す手段を返す。
// production では exec.LookPath。
//
// @ensure 戻りは exec.LookPath。
func LookPath() func(file string) (string, error) {
	return exec.LookPath
}

// CommandRun は name を起動し stdin を渡し、成功時の stdout を返す手段を返す。
// production では os/exec。Infra Adapter は os/exec を import せず、この関数結果を inject される。
//
// @ensure 成功時は stdout bytes。非 0 exit は error（stderr を原因へ含める）。
func CommandRun() func(ctx context.Context, name string, args []string, stdin []byte) (stdout []byte, err error) {
	return func(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = bytes.NewReader(stdin)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			if stderr.Len() == 0 {
				return nil, err
			}
			return nil, fmt.Errorf("%w: %s", err, stderr.String())
		}
		return stdout.Bytes(), nil
	}
}
