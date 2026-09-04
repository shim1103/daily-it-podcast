---
name: System の Cursor 認証は暫定で本番 CURSOR_API_KEY を使う
date: 2026-08-30T17:45:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Generator System workflow の Cursor 認証は、暫定で `secrets.CURSOR_API_KEY`（本番名）を `CURSOR_API_KEY` へ載せる。
2. GetX / Gemini / Google OAuth / Drive は従来どおり `TEST_*` Secrets / Variables を使う。
3. `TEST_CURSOR_API_KEY` が PR80 probe と同値で API 到達できると確認できたら、Cursor も `TEST_*` へ戻す。

## 2. Reason

1. smoke を親 env 継承・`--sandbox disabled`・絶対 path で揃えても、`TEST_CURSOR_API_KEY` では `Failed to reach the Cursor API`（run 33302114216）。
2. 同条件で `secrets.CURSOR_API_KEY` に切り替えると smoke は緑（run 33302168305）。差は Secret 名＝値の差である。
3. PR80 probe 成功条件も `CURSOR_API_KEY` だった。System を緑にする最短経路は同 Secret を Cursor だけ採用すること。

## 3. Rejected

1. `TEST_CURSOR_API_KEY` を直し終わるまで System を止める案 — probe 再現の切り分けが既に完了しており、他 vendor の TEST_* 検証を遅らせる。
2. System 全体を本番 Secrets にする案 — Drive / OAuth の test 分離契約を壊す。
