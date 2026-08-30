---
name: Gemini TTS HTTP timeout を 5 分にする
date: 2026-08-30T21:15:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `composition.geminiHTTPTimeout` を 120 秒から 5 分へ上げる。

## 2. Reason

1. run 33310692613 で draft 成功後、`Client.Timeout exceeded while awaiting headers`（120s）で落ちた。
2. TTS は長文で headers 待ちが長い。共有 HTTP の 30s より長く、かつ 120s でも足りない実測がある。

## 3. Rejected

1. 120s のまま再 dispatch する案 — 同じ上限で切れる。
2. timeout を無制限にする案 — hung request を打ち切れない。
