package cursorcli

// Cursor CLI（非SDK）の argv 構成と実行仕様に必要な確定値だけを定義する。
// 詳細な呼び出し方（Launch、parse、error変換）は Adapter が実装する（このファイルは定数のみ）。
//
// why not: 秘密値・project dir・child env allowlist・runtime 選択はここに置かない。
// それらは Composition と command launcher 実装の知識であるため、Adapter は持たない。
// ただし inject env 名（N2）は Cursor CLI の呼び出し仕様であり argv flag と同格の層であるため、ここに置く。
const (
	BinaryName = "agent"

	ModelID = "composer-2.5"

	// Cursor CLI は mode と model を独立に選べるため、mode をここで固定する。
	Mode = "ask"

	// CursorAPIKeyEnvName は Cursor CLI が API key を受け取る環境変数名である。
	// argv flag 名（--model 等）と同じ層の Cursor CLI 呼び出し仕様。
	CursorAPIKeyEnvName = "CURSOR_API_KEY"

	PrintFlag    = "-p"
	OutputFormat = "json"
	OutputFlag   = "--output-format"
	ModelFlag    = "--model"
	ModeFlag     = "--mode"

	// GHA runner が sandbox 非対応（AppArmor 未実装）のため無効化する。
	SandboxFlag  = "--sandbox"
	SandboxValue = "disabled"

	// trust は確認待ちを減らすために利用する（write ではなく実行手順の簡略化）。
	TrustFlag = "--trust"
)
