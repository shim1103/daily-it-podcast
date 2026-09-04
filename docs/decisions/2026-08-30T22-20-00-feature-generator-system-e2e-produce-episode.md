---
name: Gemini TTS 再試行 backoff を 20s 起点・上限 2m にする
date: 2026-08-30T22:20:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Gemini Adapter の retry backoff は base `20s`、exponential、上限 `2m` とする。
2. `MaxAttempts = 6` は維持する。

## 2. Reason

1. run 33313682450 / 33313642634 で Cursor 成功後に `gemini: http_status: status 429` で落ちた。
2. 旧 backoff（1s, 2s, 4s…）では rate limit が解消する前に試行が尽きる。

## 3. Rejected

1. MaxAttempts だけ増やす案 — 短い間隔の連打は 429 を悪化させる。
2. segment 間に固定 sleep を Application へ入れる案 — vendor 事情を Application が知る。
