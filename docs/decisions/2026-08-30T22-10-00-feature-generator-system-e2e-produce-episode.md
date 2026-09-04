---
name: Gemini TTS は共有 HTTP より長い Client timeout を使う
date: 2026-08-30T22:10:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. GetX / Drive / OAuth の共有 `httpTimeout` は 30s のまま。
2. Gemini TTS 専用 Client の timeout は `geminiHTTPTimeout = 120s` とする。

## 2. Reason

1. run 33310692613 で `gemini: do: ... Client.Timeout exceeded while awaiting headers`。共有 30s では長文 TTS の headers 待ちが切れる。
2. `do` は retryable だが、毎回 30s で切れて MaxAttempts を食い潰すだけになる。先に 1 回の待ち時間を延ばす。

## 3. Rejected

1. 共有 timeout を全体 120s にする案 — GetX 失敗検知が遅くなる。
2. MaxAttempts だけ増やす案 — 根本の待ち不足を直さない。
