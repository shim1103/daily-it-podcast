---
name: getEpisode JSON API 廃止と listEpisodes 原稿全文化
date: 2026-08-31T00-00-00
branch: refactor/playback-hooks-components-kiss-slim
---

## 1. Decision

1. `GET /episodes/:episodeId` の JSON 1件取得（getEpisode）を HTTP route / use-case / controller / web client から削除する
2. `listEpisodes` の各 item を `episodeItemSchema`（`body` 全文 + `audioRef`）に統一し、slim list（`topics: {title}[]` のみ）を廃止する
3. web の `select(episodeId)` は既に load した list から lookup し、2nd fetch しない
4. `GET /episodes/:episodeId/audio`（音声 byte 取得）は維持する
5. Port `getManuscript` は getEpisode 削除後に use-case から参照されなくなったため削除する（音声経路は `getAudio` Port のみ使用）
6. Request / Audio 系 identifier の `getEpisode` 接頭辞も廃止する（`EpisodeIdRequestSchema` / `getAudio` / `GetAudioController` 等へ rename。HTTP path `/episodes/:episodeId/audio` は変更しない）

## 2. Reason

1. 個人規模の episode 件数では list 1回で detail 表示に足り、get/list の二重契約と 2nd fetch が YAGNI
2. slim list と get 用 schema の二重管理が DRY 違反。1 schema に統一すると worker / web / test の射影ロジックが消える
3. `getManuscript` は getEpisode use-case 専用だった。audio 経路は wav byte 取得のみで json 1件 fetch は不要（grep で確認）

## 3. Rejected

1. getEpisode を残し web だけ list lookup — 契約と worker 実装が dead code のまま残る
2. list は slim のまま optional body — 二重契約が続き、client が list/detail で分岐を持つ
3. getManuscript を「将来用」で残す — 参照 use-case が無く、KISS に反する
