---
name: Cursor CLI 経路の結線を、each-composition file が capability config を受け取り Adapter を組む他3経路と対称にする
date: 2026-08-29T13:48:58+09:00
branch: feature/generator-composition-http-adapters
---

## 1. Decision

1. Cursor CLI 経路の Adapter 結線を `internal/composition/cursorcli.go`（each-composition file）に閉じる。`cfg.Cursor.APIKey.Reveal()` の呼び出し、Launcher 実装の選択、`SecretEnv` の構築をこの file が行う。
2. `internal/composition/runtime.go` は vendor 非依存の production runtime 既定値（`sharedHTTPClient()` とその timeout）だけを持つ。Cursor 固有の名前・factory・allowlist ラッパを持たない。
3. narrow integration test が参照する allowlist の SSoT は `internal/infrastructure/commandlaunch` の exported 定数とし、`composition` の test 都合の exported 関数を廃止する。

## 2. Reason

結論1: 直近 Decision（`docs/decisions/2026-08-29T10-50-44-feature-generator-composition-http-adapters.md`）結論4 は「Cursor CLI command 経路の結線を、他3 capability と同じ『Composition が `config.Config` の capability 別 field を Adapter へ渡す』対称形にする」と定めた。現状は `composition/cursorcli.go` が `cfg` 丸ごとを `runtime.go` の factory へ横流しし、`.Reveal()` が `runtime.go` で呼ばれるため、`gemini.go` 等が each-composition file 内で `cfg.<Cap>` を受け取り Adapter を組むのと非対称。結線を各 file へ閉じることで、5行の結線コードが同一の形になる（Orthogonality、`configuration-boundary` §3）。

結論2: `runtime.go` が Cursor 固有名（allowlist 変数、`processenvCursorCommandLauncher`、その他）を抱えるのは、D5 supersede 後の「each-composition が capability 結線を持つ」形と食い違う。vendor 非依存の `sharedHTTPClient()` と vendor 固有物を同居させると、次の読み手が「runtime.go は何の file か」を判別できない（SRP）。

結論3: `composition` が `CursorCommandInheritedEnvNameAllow()` を export する唯一の理由が narrow integration test の SSoT 参照だった。allowlist の正本が `commandlaunch` へ移る（Decision-1 結論3）ことで、test は `commandlaunch` の定数を直接参照でき、Composition Root が test 都合で公開 API を持つ状態（設計の匂い）を解消する。test が `composition` を import する逆流も消える。

## 3. Rejected

1. `.Reveal()` を `runtime.go` の共有 factory で呼び続け、指摘の他部分だけ直す案。each-composition 対称化が未達で、「Cursor だけなぜ違うのか」を後の読み手が再び問う。

2. `runtime.go` を Cursor 経路の結線 file として維持し名前だけ変える案。vendor 非依存の HTTP client 共有と Cursor 固有結線が同居し続け、変更理由が2つ混ざる。
