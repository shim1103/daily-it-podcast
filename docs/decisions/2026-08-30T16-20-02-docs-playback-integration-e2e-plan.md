---
name: playback frontend の API Client は Stub なら SU、真 HTTP だけ NI。Page BI は当面なし
date: 2026-08-30T16:20:02
branch: docs/playback-integration-e2e-plan
---

## 1. Decision

1. `fetch` を Stub / Spy した API Client test は **Sociable Unit** とする。
2. 実 network（local listener 等）へ届く API Client suite だけを **Narrow Integration** と呼ぶ。
3. Page → hooks → api / lib を全部実物にした frontend Broad Integration は **当面やらない**。配線の最終結果は認証後 E2E と下位 SU に寄せる。

## 2. Reason

1. `testing-strategy/levels.md` の Narrow は境界 provider を実際に使うこと。Stub / Spy は double であり SU と同型になる。
2. architecture `api-client.md` の Unit 観点は network-level Stub で request / Result を見る。
3. Page 合成の最終結果を E2E が既に見るなら、frontend BI は minimization に反する重複になる（generator の先送りと同型）。

## 3. Rejected

1. Stub / Spy `fetch` を Narrow と呼ぶ案 — Scope 名が実 I/O 未検証を隠す。
2. 層ごと（component–hooks、hooks–lib）に専用 Integration を今切る案 — 下位 SU と E2E で足りる範囲を超える。
