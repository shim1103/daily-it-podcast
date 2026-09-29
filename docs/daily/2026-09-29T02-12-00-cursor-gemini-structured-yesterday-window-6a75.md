---
name: 昨日 Fetch 窓・Gemini structured・日本語 topic.title prompt
date: 2026-09-29T02:12:00
session_id: cloud-6a75
branch: cursor/gemini-structured-yesterday-window-6a75
prev: なし
---

## 1. Summary

Fetch を表示 TZ の昨日 half-open 暦日へ切り替え、Gemini TextWriter だけに JSON structured 拘束を載せ、brief prompt で topic.title の日本語必須を強化した。

## 2. Changes

1. `go test`（constants / application / geminiapi / composition / manuscript / test）緑
2. cron は未変更

### Commits

1. `f01ad81` — feat(generator): Fetch 窓を表示 TZ の昨日 half-open 暦日へ切り替える
2. `b722a1f` — feat(generator): Gemini TextWriter だけに WriterOutput JSON schema 拘束を載せる
3. `99c8370` — docs(generator): brief prompt で topic.title の日本語必須を強化する
