package cursorcli

// Cursor CLI（非SDK）の argv 構成に必要な確定値だけを定義する。
// 詳細な呼び出し方（Launch、parse、error変換）は Adapter が実装する（このファイルは定数のみ）。
//
// why not: 秘密名・project dir・child env allowlist・runtime 選択はここに置かない。
// それらは Composition と command launcher 実装の知識であるため、Adapter は持たない。
const (
	BinaryName = "agent"

	ModelID = "composer-2.5"

	// Cursor CLI は mode と model を独立に選べるため、mode をここで固定する。
	Mode = "ask"

	PrintFlag    = "-p"
	OutputFormat = "json"
	OutputFlag   = "--output-format"
	ModelFlag    = "--model"
	ModeFlag     = "--mode"

	// sandbox を有効化して、誤った write を隔離する。
	SandboxFlag  = "--sandbox"
	SandboxValue = "enabled"

	// trust は確認待ちを減らすために利用する（write ではなく実行手順の簡略化）。
	TrustFlag = "--trust"
)
