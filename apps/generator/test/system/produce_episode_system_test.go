//go:build system

// Scope: System
// 実物: cmd/generator 入口から出口までの合成経路（振る舞い実装は後続）
// Double: なし（credential 付き実 operation。実行場所の正は Decision を参照）
// @require build tag `system` 付きで本 package だけが収集される。Integration gate 入口は本 package を実行しない。
// @ensure （後続 System Issue が実装する）subprocess 入口成功時に契約上の最終成果物が残る。
// @invariant local に secret を置かない。本番 credential を使わない。
package system
