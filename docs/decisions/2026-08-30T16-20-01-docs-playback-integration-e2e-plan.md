---
name: playback Unit coverage 分母は SU と secret なし Narrow の合算
date: 2026-08-30T16:20:01
branch: docs/playback-integration-e2e-plan
---

## 1. Decision

1. playback Unit coverage の計測分母に、**secret なし Narrow Integration** を含める（Sociable Unit + Narrow の合算）。
2. Broad Integration 以上と browser E2E は分母に入れない。
3. 閾値の数値変更は本 Decision の対象外とする。正本は `apps/playback/vitest.config.mjs` と `DESIGN.md`。

## 2. Reason

1. `testing-strategy/coverage.md` は外部境界単位を Unit と Narrow の両方で見る。SU / NI を分離すると実 I/O 行は SU だけでは埋まらない。分母から外すと検出力がすり抜ける（generator `2026-08-30T11-52-00` と同型）。
2. Broad / E2E は結線・最終結果が目的であり、行 coverage の逃げ道にしてはならない（Pyramid）。

## 3. Rejected

1. Broad を coverage 分母に入れる案 — 上位 Scope で下位の穴を埋める運用になる。
2. Narrow 対象を除外リストへ逃がす案 — 実ロジックがある package を除外方針に載せると gate が弱くなる。
