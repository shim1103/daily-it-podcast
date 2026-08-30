---
name: Playback e2e 配置・安定fixture・実spec を deploy 前に固定する
date: 2026-08-31T00:57:00
session_id: none
branch: feature/playback-e2e-deploy
prev: なし
---

## 1. Summary

Playback の Integration/e2e 配置方針と本番 Drive 安定 fixture（ProduceEpisode/TextWriter 検証込み）、認証後 Playwright 実 spec（Secret 未設定は skip）を Decision と code に固定した。Deploy Phase 1–2 と storageState 付き e2e 緑は未実施。

## 2. Changes

1. Decision 3 本で test/integration・test/e2e 分割、安定 fixture、ProduceEpisode 検証範囲を固定した。
2. Integration 4 suite を `test/integration/` へ移し vitest glob を更新した。
3. placeholder をやめ、一覧・原稿・audio の remote e2e と安定 fixture artifact を追加した。Secret 無しで 3 skip / exit 0 を確認した。
4. lane / e2e 達成契約を Decision と実 spec に合わせた。

### Commits

- `07d572b`
- `1e8cd54`
- `96ab099`
- `581c4d6`

## 3. Append

1. PR #105 を作成した（https://github.com/shim1103/daily-it-podcast/pull/105）。base `develop`。`origin/develop` へ rebase して conflict を解消した。
