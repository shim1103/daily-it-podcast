---
name: Broad Integration は production Adapter 結線と同型の手組みで足り FromEnv を通さない
date: 2026-08-30T15:42:00
branch: feature/generator-broad-integration-produce-episode
---

## 1. Decision

1. secret なし Broad Integration は、Composition Root の **production Adapter 型・順序**と同型の結線を test 内で組み立て、真外部だけを httptest TLS redirect / fake child 等で double する。
2. Broad は `NewProduceEpisodeFromEnv` および Composition の `sharedHTTPClient` / `sharedLookupEnv` / `config.Load` 経路を **通さない**。
3. env 読取・Config 正規化・CLI Driving Adapter を含む入口全体の検証は System（`cmd/generator` subprocess）の責務とし、Broad に混ぜない。
4. 収集境界・gate の正は先行 Decision（`docs/decisions/2026-08-30T11-56-00-docs-generator-broad-system-e2e-plan.md` / `2026-08-30T11-56-01`）を参照する。本 Decision は Broad の結線境界だけを固定する。

## 2. Reason

1. Broad が所有するのは「複数 production 実装の配線・状態伝播・error 伝播」である（`testing-strategy/levels.md`）。`FromEnv` は env → Config → 結線の **entrypoint 関数**であり、Adapter 合成グラフそのものではない。
2. secret なし Broad は DialTLS 等で真外部を差し替える必要がある。`sharedHTTPClient()` は注入口を持たないため、`FromEnv` を直呼びすると double 不能になる。同型 Adapter を custom `*http.Client` で手組みする方が、Broad の観測対象を壊さずに Repeatable を保てる。
3. `FromEnv` 直呼びを System と呼ぶ案は既に Rejected（`docs/decisions/2026-08-30T11-56-01`）。Broad で同経路を半端に通すと、Scope 名と入口契約が再び混ざる。

## 3. Rejected

1. Broad から `NewProduceEpisodeFromEnv` を直呼びする案 — DialTLS 注入不能で真外部 double ができず、secret なし gate と両立しない。
2. Composition に test 用 DialTLS hook を足して `FromEnv` を Broad から通す案 — production Composition に test 専用分岐を持ち込み、Composition が「結線」以外の関心を抱える。
3. Broad で env 読取・`config.Load` 失敗まで assert する案 — Narrow / Unit / System が所有する境界の再 assert になり、minimization に反する。
