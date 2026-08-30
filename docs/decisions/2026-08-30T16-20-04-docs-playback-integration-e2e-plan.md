---
name: Playback deploy 進捗は 3 Phase（前ゲート / deploy+Verification / 運用後続）
date: 2026-08-30T16:20:04
branch: docs/playback-integration-e2e-plan
---

## 1. Decision

Playback deploy の進捗 Phase は次の **3** とする。

1. deploy 前ゲート
2. 初回 `wrangler deploy` + `DEPLOY.md` §7 browser Verification（OTP・一覧・再生）
3. 運用後続（rollback 文書化等）

旧「検証専用 Phase」と「運用後続 Phase」を分けた 4 Phase 構成は採らない。進捗 index の正は `docs/tasks/todo/playback-lane.md`、Verification 項目の正は `DEPLOY.md`。

## 2. Reason

Verification を独立 Phase にすると、deploy 完了と受け入れ完了が二重進捗になり、lane と DEPLOY の参照がずれる。deploy 直後に §7 を同じ Phase で閉じる方が完了定義が一意になる。

## 3. Rejected

1. 4 Phase（検証と運用後続を分離）のまま残す案 — 進捗と完了定義が二重になる。
