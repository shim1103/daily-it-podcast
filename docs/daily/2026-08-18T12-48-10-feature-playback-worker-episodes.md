---
name: playback worker の Episode 読取 UseCase と Fake Drive
date: 2026-08-18T12:48:10
session_id: none
branch: feature/playback-worker-episodes
prev: なし
---

## 1. Summary

`origin/develop` から `feature/playback-worker-episodes` を切り、worker Application（Port / List / Get JSON / Get 音声）と in-memory Fake Drive adapter を実装した。Unit 27 件と Integration（0 file）が pass した。実 Google Drive adapter は未切り出しとして lane に残し、episodes todo を削除した。

## 2. Changes

1. `EpisodeRepository` Port と `listEpisodes` / `getEpisode` / `getEpisodeAudio` UseCase を追加
2. `InMemoryEpisodeRepository` と manuscript schema 検証を追加（Fake Drive）
3. `playback-worker-episodes` todo を削除し、`playback-lane.md` に Fake 完了と実 Drive 未切り出しを追記

### Commits

- `bd7139a` — feat(playback): worker の Episode 読取 Port と UseCase を追加する
- `53e0a02` — feat(playback): Drive 読取の in-memory Fake adapter を追加する
- `d89259f` — docs(playback): worker-episodes 完了と実 Drive 未切り出しを lane に残す
