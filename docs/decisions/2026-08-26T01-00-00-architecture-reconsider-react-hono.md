---
name: Go/Node の toolchain version は go.mod と .nvmrc を正本にし、GitHub Actions は file参照で追従する
date: 2026-08-26T01:00:00
branch: docs/architecture-reconsider-react-hono
---

## 1. Decision

1. Go version の正本は `apps/generator/go.mod` の `go` directive。`toolchain` directiveは追加しない
2. Node version の正本は新設する `apps/playback/.nvmrc`
3. `.github/workflows/test-unit.yml` / `test-integration.yml` の `actions/setup-go` は `go-version-file: apps/generator/go.mod`、`actions/setup-node` は `node-version-file: apps/playback/.nvmrc` を使い、version文字列をYAMLへ直書きしない
4. local向けの安全網として `apps/playback/package.json` に `engines.node`、`apps/playback/.npmrc` に `engine-strict=true` を追加する

## 2. Reason

1. 導入前は `go.mod` の `go 1.26.6` と両 workflow YAML の `go-version: "1.26.6"` に同じ値が3箇所直書きされ、Node側も両YAMLの `node-version: "22"` が2箇所に直書きされていた。DRY（`design-philosophy.md §2-2`）に反する重複であり、version更新時に一部を更新し忘れるriskがある
2. `go-version-file` / `node-version-file` は `actions/setup-go` / `actions/setup-node` 公式が提供する標準optionであり、値の参照元をrepo内の既存fileに一本化できる。特殊なhackではない
3. `toolchain` directiveは Go 1.21以降の標準機構で、`go`コマンド自体が`go.mod`の指定versionより古い場合に自動でtoolchainを取得する。ただし`go` directiveと同一versionを`toolchain`に明記しても`go mod tidy`が自動的に除去する（Goが冗長と判定するため）。このrepoでは`go` directiveの値がそのままlocal/CIの要求versionとして機能するため、`toolchain` directiveの追加は不要
4. `.nvmrc`はNode生態系で最も広く使われる規約（tool非依存のplain text）。Voltaのような専用tool導入は開発者全員へのtool install強制を伴い、個人開発の現規模には過剰（Rule of Least Power、`design-philosophy.md §4-2`）
5. `.nvmrc`・`go.mod`はどちらも「書いてあるだけ」では強制力を持たない。localで実際に異なるversionを使ってしまうケースに備え、`engines` + `engine-strict=true`で`npm ci`/`npm install`時点のfail-fastを追加する。この安全網は実際に動作確認済み（local Node 26 で `.nvmrc` の 22 と不一致の状態から `npm ci` が `EBADENGINE` で失敗することを確認した）

## 3. Rejected

1. Volta導入 — 自動切替の体験は強いが、開発者全員へのVolta install前提を追加する。個人開発の現規模でこの依存追加はRule of Least Powerに反する
2. `.nvmrc`を置くだけで`engines`を追加しない案 — `.nvmrc`はtoolが明示的に読まない限り強制力がゼロ。`npm ci`実行時点で気づける安全網が無いまま放置するのは、Node versionズレの根本原因を再発させる
3. `go.mod`へ`toolchain go1.26.6`を明記する案 — `go` directiveと同一versionの`toolchain`明記は`go mod tidy`が冗長と判定し自動的に除去する。Goの仕様上維持できない
