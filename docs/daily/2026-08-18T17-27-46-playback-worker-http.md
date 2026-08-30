---
name: playback-worker-http の HTTP 境界実装と todo 整備
date: 2026-08-18T17:27:46
session_id: none
branch: feature/playback-worker-http
prev: 2026-08-18T14-18-00-feature-playback-web.md
---

## 1. Summary

`apps/playback` の worker HTTP 境界を Route / Controller / External Error 写像で分離実装し、未知失敗は契約の `code` へ写像する形に整えた。音声成功時は `Content-Type: audio/wav` 相当で raw byte を返すように修正した。あわせて `docs/tasks/todo` の lane と HTTP refactor 予定を整備し、不要になった `playback-worker-http.md` は削除済みにした。

## 2. Changes

1. `apps/playback` に Route / Controller / logging / error mapping / audio byte 作法を追加
2. `cd apps/playback && npm run test:unit` と `test:integration` を再実行し、unit は pass（integration は no tests）
3. `docs/tasks/todo/playback-lane.md` の依存表更新、HTTP refactor todo の追加、既存 `playback-worker-http.md` の削除

### Commits

1. `75560c8` — feat(playback): HTTP 境界の Route/Controller/Mapping 実装
2. `c3cc943` — docs(tasks): playback lane と HTTP refactor todo の整備

