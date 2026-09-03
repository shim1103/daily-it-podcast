# chore(ci): GitHub Actions を Node24 world へ更新する

## 1. Summary

このIssueでは、全 workflow の `actions/checkout` と `actions/setup-go`（および `actions/setup-node`）を Node24 runtime の major version へ上げ、runner の Node20 deprecation warning を解消する。完了後、`generator-*` / `test-*` の各 run のログに "Node.js 20 is deprecated" の警告が出ない。

## 2. Context

事実:
1. run 33760360649（`generator-produce-episode.yml` の workflow_dispatch）で以下の warning が出た。
   `Warning: Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: actions/checkout@v4, actions/setup-go@v5.`
2. `.github/workflows/` 6 file 全てが `actions/checkout@v4` + `actions/setup-go@v5` を使う。`test-unit.yml` と `test-integration.yml` は加えて `actions/setup-node@v4` を使う。
3. warning のみで run は成功する。runner が Node24 へ強制昇格させて実行している。

仮定:
1. `actions/checkout@v5` / `actions/setup-go@v6` / `actions/setup-node@v5` が Node24 target で、機能的な破壊的変更はこの repo の使い方（`go-version-file` / `node-version-file` / `cache`）には及ばない。実装時に各 release note で確認する。

## 3. Canonical Sources

1. `.github/workflows/*.yml` — 対象 6 file。各 file 冒頭の `@require` / `@invariant` comment が toolchain 前提の SSOT。
2. GitHub Changelog "Deprecation of Node 20 on GitHub Actions runners"（2025-09-19）— warning の出典。
3. `apps/generator/go.mod` / `apps/playback/.nvmrc` — Go / Node version の SSOT。この Issue で変更しない。

## 4. Scope

### In Scope

1. 6 workflow の `actions/checkout@v4` → 新 major。
2. 6 workflow の `actions/setup-go@v5` → 新 major。
3. `test-unit.yml` / `test-integration.yml` の `actions/setup-node@v4` → 新 major。
4. 各 action の `with:` block が新 major で有効かの確認と、必要なら key 名の追随。

### Out of Scope

1. Go / Node 自体の version 変更（`go.mod` / `.nvmrc`）。
2. workflow の step 構成・trigger・job 分割の変更。
3. `golangci-lint` の version（`test-unit.yml` の curl install）。

## 5. Contract

- workflow の trigger（`on:`）、job 名、step 名、実行する script は不変。
- `go-version-file: apps/generator/go.mod` / `node-version-file: apps/playback/.nvmrc` / `cache: npm` の指定は維持する。

## 6. Constraints

1. 6 file すべてで同じ major version に揃える（Principle of Least Astonishment §4-5 / §5-1）。1 file だけ別 version にしない。
2. version は major pin（`@v5` 形式）にする。full-length SHA pin へ切り替えるのはこの Issue の scope 外。

## 7. Acceptance Criteria

- [ ] `.github/workflows/*.yml` 6 file に `actions/checkout@v4` / `actions/setup-go@v5` / `actions/setup-node@v4` が残っていない。
- [ ] 6 file の `actions/checkout` の major version が一致する。`actions/setup-go` も一致する。
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
- `shim gh create-issue` での起票と Project 追加は別途行う（この task file 作成時点では未実行）。
