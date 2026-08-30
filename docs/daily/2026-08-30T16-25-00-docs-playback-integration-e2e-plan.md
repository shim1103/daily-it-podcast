---
name: Playback Integration/E2E の A/B 固定と C 達成契約
date: 2026-08-30T16:36:00
session_id: none
branch: docs/playback-integration-e2e-plan
prev: なし
---

## 1. Summary

Playback の NI/BI/E2E について Decision を問い単位で固定し、E2E 入口（Playwright・workflow・script）と C 達成契約4本まで到達した。振る舞い本体（C 実装）は未着手。

## 2. Changes

1. Decision 5本で gate/E2E 定時・coverage・frontend Scope・Access/storageState・3 Phase を固定した。
2. DESIGN/DEPLOY/lane を揃え、`PLAYWRIGHT_*` の登録・取得・GHA 写像を DEPLOY に正本化した。
3. Vitest 収集・Playwright placeholder・`playback-e2e.yml`・gate composer 隔離検査を置いた。
4. C 達成契約4本を lane に登録した。

### Commits

- `ecf698d`
- `3ed1050`
- `160cccd`
- `93df5a8`
