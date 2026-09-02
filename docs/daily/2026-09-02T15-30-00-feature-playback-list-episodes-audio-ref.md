---
name: playback web 直交 Decision と A 契約・C task 化
date: 2026-09-02T15-30:00
session_id: none
branch: feature-playback-list-episodes-audio-ref
prev: 2026-08-30T03-19-00-feature-playback-list-episodes-audio-ref
---

## 1. Summary

playback web の selection と playback 直交を domain 固有 Decision 1 本に固定し、A 契約（型・stub・SU test 35 件）を追加した。C を view-models / ui-rewrite / legacy-cleanup の 3 task に再編し、旧 `playback-audio-player-ui-design.md` を置換した。

## 2. Changes

1. 誤った束ね Decision（orchestration）を削除し、SRP 準拠の orthogonality Decision へ差し替えた。
2. `playback-state`、hash adapter、hook stub 5 本、component 契約 3 本と各 SU test を追加。`test:unit` 257 passed。
3. lane と C task 3 本を更新。lessons に Decision 分割と A 時点 test 化の知見を追記。

### Commits

- `4ce917a`
- `5095714`
- `772e97d`
- `4abecf7`
