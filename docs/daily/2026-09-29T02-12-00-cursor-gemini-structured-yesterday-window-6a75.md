---
name: 昨日 Fetch 窓・Gemini structured・日本語 topic.title prompt
date: 2026-09-29T02:12:00
session_id: cloud-6a75
branch: cursor/gemini-structured-yesterday-window-6a75
prev: なし
---

## 1. Summary

Fetch を表示 TZ の昨日 half-open 暦日へ切り替え、Gemini TextWriter だけに JSON structured 拘束を載せた。査読後、until は Adapter、schema は WriterOutput 型から生成、bookend は openingIntro/endingSummary に対称化した。

## 2. Changes

1. `go test ./internal/... ./test/` 緑
2. cron は未変更
3. lessons 追記・修正

### Commits

1. `f01ad81` — feat(generator): Fetch 窓を表示 TZ の昨日 half-open 暦日へ切り替える
2. `b722a1f` — feat(generator): Gemini TextWriter だけに WriterOutput JSON schema 拘束を載せる
3. `99c8370` — docs(generator): brief prompt で topic.title の日本語必須を強化する
4. `ec18349` — refactor(generator): Fetch 窓の until 判定を Adapter へ移す
5. `9576f38` — refactor(generator): WriterOutput schema を contracts 経由で Gemini へ渡す
6. `bd3ce33` — test(generator): brief prompt test に余計な文言 assert 禁止を declare する
7. `5695527` — refactor(generator): WriterOutput schema 正本を entities/models へ移す
8. `5e699d1` — test(generator): ItemSource Adapter に until 境界 SU を追加する
9. `0778fa3` — docs(log): WriterOutput models 正本と until 境界 test の lessons を直す
10.  — refactor(generator): WriterOutput から schema 生成し bookend field を対称化する
11.  — docs(generator): WriterOutput 型生成 schema と bookend rename の Decision を残す
10.  — refactor(generator): WriterOutput から schema 生成し bookend field を対称化する
11.  — docs(generator): WriterOutput 型生成 schema と bookend rename の Decision を残す
