---
name: 初回 deploy/e2e 完了後の運用 SSOT と lane 整理
date: 2026-08-31T02:19:00
session_id: none
branch: drafts-playback-ops-ssot
prev: 2026-08-31T00-57-00-feature-playback-e2e-deploy.md
---

## 1. Summary

Playback 初回 deploy・Access・認証済み e2e（GHA success）まで到達したあと、完了達成契約を削除し lane を運用残へ縮め、DEPLOY.md を継続運用 SSOT に書き換えた。master 直編集は policy 拒否のため drafts PR 経由。

## 2. Changes

1. 本番 Worker は Access 入場後に一覧・原稿・再生できた。CLIENT_ID Variable の誤配置（Secret 形）を直し `/episodes` の 503 を解消した。
2. `storageState` を取得し GHA Secret を登録した。`playback-e2e` を develop で dispatch し success（run 33324315110）。
3. develop↔master 無共通祖先のため PR #106 で unrelated histories 合流し、default branch に `playback-e2e.yml` を載せた。
4. 達成契約 2 file を削除し lane / DEPLOY を運用向けに更新した。PR #107。

### Commits

- `6854a30`
- `1520166`
