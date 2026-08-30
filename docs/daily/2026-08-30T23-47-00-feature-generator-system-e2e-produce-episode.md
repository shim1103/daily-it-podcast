---
name: Generator System e2e 緑化の途中停止（Gemini 429）
date: 2026-08-30T23:47:00
session_id: none
branch: feature/generator-system-e2e-produce-episode
prev: なし
---

## 1. Summary

System suite は Cursor 経路まで通せる状態まで進めたが、Gemini TTS の 429（quota）で dispatch 緑化は未達。同日 skip / Drive upsert / 公開型書込の Decision と実装、Gemini retry・timeout 延長までを branch に載せた。

## 2. Changes

- Cursor smoke は本番 `CURSOR_API_KEY` で緑。`TEST_*` は値不一致で失敗しやすい。
- System 失敗の主因は Gemini `status 429`（例 run `33314746860`）。draft 到達後に TTS で止まる。
- 同日完成ペアの Fetch 前 skip、Drive 同 stem upsert、json→wav 公開順を Decision 化し Application に実装。
- Gemini: Retry-After・callGap・backoff 60s/3m、System job/test timeout 延長を push。当該版の dispatch 検証は未実施。
- 引き継ぎと日quota memo は `docs/tasks/todo/generator-system-e2e-produce-episode.md`。

### Commits

- `b93cc6c`
- `fbfd330`
- `79845dd`
- `0a5993c`
