---
name: getEpisode JSON API 廃止と list 原稿全文化
date: 2026-08-31T00:37:24
session_id: none
branch: refactor/playback-hooks-components-kiss-slim
prev: なし
---

## 1. Summary

playback の 1件 JSON API を消し、`listEpisodes` が原稿全文と `audioRef` を返す形へ契約を統一した。音声 GET は `getAudio` へ rename し、生きた identifier / comment / todo から `getEpisode` 字列を落とした。歴史記録（decisions / daily）は触っていない。

## 2. Changes

1. ask/plan で個人規模の payload を許容し削除方針を凍結した。
2. executor 実装後、manager が test/typecheck/lint と grep で独立確認した。
3. `GetEpisodeRequest` 残置と `getEpisode` 字列 NG の指摘を受け、Request / Audio 系を rename した。
4. 生きた todo 文言を直し、decisions/daily は歴史として残した。

### Commits

- `225e8f0`
- `7a13701`
- `d25577d`
- `02bb3c6`
- `4944ed1`
