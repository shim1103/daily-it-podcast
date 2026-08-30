// Package commandlaunch は command 起動の契約と、
// child process へ渡す秘密環境変数の入力型を提供する。
// SecretEnv.Value を含む秘密値は error / log / stdin へ写さない。
// InheritedEnvNameAllow() は POSIX child process 一般の最小継承環境変数名の
// コピーを返す。特定 launcher 実装や vendor に依存しない。
package commandlaunch

import "context"

// inheritedEnvNameAllow は child process へ親環境から継承を許す環境変数名の SSoT である。
// PATH、HOME、TMPDIR は全 CLI 実行に必要な最小限の環境変数であり、実装や vendor に依存しない。
var inheritedEnvNameAllow = [...]string{"PATH", "HOME", "TMPDIR"}

// InheritedEnvNameAllow は child process へ親環境から継承を許す環境変数名のコピーを返す。
// PATH / HOME / TMPDIR は全 CLI 実行に必要な最小限の環境変数であり、実装や vendor に依存しない。
// 戻りはコピーであり、呼び出し側が変更しても SSoT を壊さない。
func InheritedEnvNameAllow() []string {
	return append([]string(nil), inheritedEnvNameAllow[:]...)
}

// Command は起動する command の秘密値を含まない入力である。
// Args と Stdin は秘密値を含まない。
type Command struct {
	Program string
	Args    []string
	Stdin   []byte
}

// SecretEnv は child process へ渡す1つの秘密環境変数のエントリである。
// Name は環境変数の名前であり、Value は秘密値である。
// Value は error / log / stdin へ写してはならない。
type SecretEnv struct {
	Name  string
	Value string
}

// Launcher は command を起動する。
type Launcher interface {
	// Launch は command を起動し、成功時は stdout bytes を返す。
	// Program が空なら起動前に error を返す。
	// error には秘密値と stdin を含めない。
	// 失敗時は child stderr 先頭最大 300 byte を診断として載せてよい。
	// stderr に秘密値または stdin 本文が含まれる場合は head を載せない。
	Launch(ctx context.Context, command Command) ([]byte, error)
}

// SecretEnvLauncherFactory は inject env 名を受け取り、
// secret 値と親環境アクセス手段を閉じ込めた Launcher を返す。
// secret 値は factory の closure に閉じ、呼び出し側（vendor Adapter）へは渡らない。
type SecretEnvLauncherFactory func(envName string) Launcher
