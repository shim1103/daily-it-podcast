# chore(ci): GitHub Actions を Node24 world へ更新する

## 1. Summary

このIssueでは、全 workflow の `actions/checkout` と `actions/setup-go`（および `actions/setup-node`）を Node24 runtime の major version へ上げ、runner の Node20 deprecation warning を解消する。完了後、`generator-*` / `test-*` の各 run のログに "Node.js 20 is deprecated" の警告が出ない。

## 2. Context

事実:
1. run 33760360649（`generator-produce-episode.yml` の workflow_dispatch）で以下の warning が出た。
   `Warning: Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: actions/checkout@v4, actions/setup-go@v5.`
2. `.github/workflows/` 全 file が `actions/checkout@v4` を使う。Go 系は `actions/setup-go@v5`、Node 系（`test-unit` / `test-integration` / `playback-e2e`）は `actions/setup-node@v4` を使う。
3. warning のみで run は成功する。runner が Node24 へ強制昇格させて実行している。

仮定:
1. `actions/checkout@v5` / `actions/setup-go@v6` / `actions/setup-node@v5` が Node24 target（各 `action.yml` の `using: node24` で確認済み）。機能的な破壊的変更はこの repo の使い方（`go-version-file` / `node-version-file` / `cache`）には及ばない。

## 3. Canonical Sources

1. `.github/workflows/*.yml` — 対象全 file。各 file 冒頭の `@require` / `@invariant` comment が toolchain 前提の SSOT。
2. GitHub Changelog "Deprecation of Node 20 on GitHub Actions runners"（2025-09-19）— warning の出典。
3. `apps/generator/go.mod` / `apps/playback/.nvmrc` — Go / Node version の SSOT。この Issue で変更しない。

## 4. Scope

### In Scope

1. 全 workflow の `actions/checkout@v4` → `@v5`（Node24）。
2. Go 系 workflow の `actions/setup-go@v5` → `@v6`（Node24）。
3. `test-unit.yml` / `test-integration.yml` / `playback-e2e.yml` の `actions/setup-node@v4` → `@v5`（Node24）。
4. 各 action の `with:` block が新 major で有効かの確認と、必要なら key 名の追随。

### Out of Scope

1. Go / Node 自体の version 変更（`go.mod` / `.nvmrc`）。
2. workflow の step 構成・trigger・job 分割の変更。
3. `golangci-lint` の version（`test-unit.yml` の curl install）。

## 5. Contract

- workflow の trigger（`on:`）、job 名、step 名、実行する script は不変。
- `go-version-file: apps/generator/go.mod` / `node-version-file: apps/playback/.nvmrc` / `cache: npm` の指定は維持する。

## 6. Constraints

1. 全 file で同じ major version に揃える（Principle of Least Astonishment §4-5 / §5-1）。1 file だけ別 version にしない。
2. version は major pin（`@v5` 形式）にする。full-length SHA pin へ切り替えるのはこの Issue の scope 外。

## 7. Acceptance Criteria

- [x] `.github/workflows/*.yml` に `actions/checkout@v4` / `actions/setup-go@v5` / `actions/setup-node@v4` が残っていない。
- [x] 全 file の `actions/checkout` の major version が一致する。`actions/setup-go` / `actions/setup-node` もそれぞれ一致する。
- [ ] いずれかの workflow を実行した run のログに "Node.js 20 is deprecated" の warning が出ない。
- [ ] `test-unit` / `test-integration` の gate（static / Unit / race / integration）が緑のまま。
- [ ] `generator-produce-episode` の workflow_dispatch が成功する（従来どおり）。

## 8. Verification

```bash
# 対象が残っていないこと
grep -rn "checkout@v4\|setup-go@v5\|setup-node@v4" .github/workflows/

# 揃っていること
grep -rn "checkout@\|setup-go@\|setup-node@" .github/workflows/
```

push で `test-unit` / `test-integration` が走るので、その run のログで warning 消失と gate 緑を確認する。`generator-produce-episode` は手動 dispatch で確認する（無料枠 quota に注意し、必要なときだけ）。

## 9. Dependencies

なし。`feature/generator-system-e2e-produce-episode` の作業とは独立。

## 10. Risks

1. 新 major で `with:` の key 名や default が変わり、cache が効かない / Go version 解決が変わる可能性。→ 実装時に各 release note を読み、run ログで `go version` と cache hit を確認する。
2. `actions/checkout@v5` の default 挙動変更（過去に `fetch-depth` や submodule 周りで発生履歴あり）。→ この repo の workflow は shallow clone 前提の script しか呼ばないため影響は小さいが、run で確認する。

## 11. Notes

- warning の初出は run 33760360649。
- SHA pin による supply-chain 対策強化は別 follow-up 候補（この Issue の scope 外）。
- Node24 最小 major: `checkout@v5` / `setup-go@v6` / `setup-node@v5`（各 `action.yml` の `using: node24`）。
- local grep AC は満たした。run ログ AC は push 後に確認。
