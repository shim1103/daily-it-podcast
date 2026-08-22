// Package commandlaunch は秘密値を渡さない command 起動の契約を提供する。
package commandlaunch

import "context"

// Command は起動する command の秘密値を含まない入力である。
// Args と Stdin は秘密値を含まない。
type Command struct {
	Program string
	Args    []string
	Stdin   []byte
}

// Launcher は command を起動する。
type Launcher interface {
	// Launch は command を起動し、成功時は stdout bytes を返す。
	// Program が空なら起動前に error を返す。
	// error には秘密値、stdin、child stderr を含めない。
	Launch(ctx context.Context, command Command) ([]byte, error)
}
