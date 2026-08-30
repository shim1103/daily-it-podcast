---
name: listEpisodes に audioRef を載せて API だけ PR 化する
date: 2026-08-30T03:19:00
session_id: none
branch: feature/playback-list-episodes-audio-ref
prev: なし
---

## 1. Summary

listEpisodes の各 item に `audioRef` を必須化した。fake 原稿の topic1 `startSec` を opening 後へずらした。web 既存 fixture を契約へ追随させた。UI/audio refactor は未完成のため別 branch に WIP 退避し、本 branch の PR 対象外とした。

## 2. Changes

1. PR scope は API + fixture + decision のみ。UI decision 3本と Feature 分割は非対象。
2. UI 途中差分は `feature/playback-audio-player-ui-design-api-refactoring` 上の `wip(playback)` commit に残した。
3. unit 256 / typecheck は API-only branch で green。

### Commits

- `d577625`
- `1622fd3`
- `b0642c7`
