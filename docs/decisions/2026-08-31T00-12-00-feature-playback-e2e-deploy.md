---
name: Playback の Integration 以上は test/integration、browser E2E は test/e2e。contract 専用 dir は置かない
date: 2026-08-31T00:12:00
branch: feature/playback-e2e-deploy
---

## 1. Decision

1. `apps/playback/test/` 配下の実行 suite は **`integration/`** と **`e2e/`** の2 dir に分ける。
2. Narrow / Broad Integration（Vitest・gate 内）は `test/integration/` に置く。browser E2E（Playwright・gate 外）は `test/e2e/` に置く。
3. **contract 専用 directory は作らない。** Contract が必要になった場合も、分類は file 名（`*contract*`）で出し、dir 名で補完しない（先行の命名規約を維持）。
4. 収集 glob・runner 設定の正本は `apps/playback/vitest.config.mjs` と `apps/playback/playwright.config.ts` とする。本 Decision に glob 文字列を写して正本化しない。

## 2. Reason

1. Integration と E2E は runner・gate・credential 前提が違う。同じ `test/` 直下に平置きすると、収集境界と失敗原因の切り分けが path から読めない。
2. `test/contract/` を先に切ると、CDC 未着手の空領域が増え、minimization / YAGNI に反する。分類語 `contract` は file 名で足りる。
3. 先行 Decision（`2026-08-30T16-20-00`）は gate と分類語を決めたが、directory 分割までは固定していない。配置の再発問いに答えるのが本 file の範囲である。

## 3. Rejected

1. `test/` 直下へ Integration を平置きし続ける案 — runner 差が path から読めず、e2e 以外の「何の Integration か」が dir で弱い。
2. `test/contract/` を Integration / E2E と並列に先置きする案 — 未使用の第3領域を作り、今回の配置主張（integration / e2e）と無関係な taxonomy を増やす。
3. E2E を `web/` 配下へ移す案 — Playwright `testDir` と Integration 領域の対称性が壊れ、gate 外入口の所在が散る。
