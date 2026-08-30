---
name: Playback deploy・Access の境界契約と運用 SSOT を固定する
date: 2026-08-25T18:13:45
session_id: none
branch: feature/playback-worker-deploy
prev: なし
---

## 1. Summary

Playback Worker の初回手動 deploy に向けて、同一 origin・Access・secret の方針と wrangler 境界契約を固定した。運用 latest は `DEPLOY.md`、Reason/Rejected は Decision Record、未着手は lane の C/D に分けた。

## 2. Changes

1. QuestionPlan で Access / deploy の未決を洗い、同一 origin・`*.workers.dev`・メール OTP・session 30d・all secret・preview 禁止・deploy 実行は後続、を確定した。
2. A として `wrangler.jsonc` と薄い `worker-entry` を置き、B として `DEPLOY.md` と Decision 2 本を置いた。薄い Decision を指摘され、文書分業と Access/secret を分離し Reason を厚くした。
3. README / DESIGN から Access・OTP 等の運用本文を除き `DEPLOY.md` 参照へ寄せた。
4. `playback-lane` に deploy 前 C と後続 D を短く残した。Issue file は作っていない。
5. shared skill 側で logging/decisions の書き方と scope-split の 4 種 SSOT 分けを通るよう直した（本 repo 外）。
6. PR 前に `origin/develop` を merge し、UI 完了と deploy A/B の lane conflict を両意図で解消した。`worker-entry` unit と playback unit は merge 後も pass。
7. PR #54 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `e9e1b37`
- `ee14d93`
- `08e0b86`
- `c2422cd`
- `34d1c7a`
- `385ba7f`
- `d4001e9`
