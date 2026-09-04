---
name: Gemini MaxAttempts を 6 にし detail 下限を 48 秒へ
date: 2026-08-30T20:35:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `gemini.MaxAttempts` を 4 から 6 へ上げる。
2. `DraftTopicDetailMinSec` を 50 から 48 へ下げる（文字数 336）。Target / Max は変えない。

## 2. Reason

1. run 33308282246: Cursor〜draft 成功後、`gemini: decode_pcm: output audio is missing`（retryable）で落ちた。現行 4 回では足りない実測。
2. run 33308073574: `topic[1].detail rune count 348 is out of range [350, 770]`。下限を 2 秒分だけ下げて LLM の惜しい不足を吸収する。

## 3. Rejected

1. 再 dispatch だけで済ませる案 — 同失敗が連続している。
2. detail 下限を大きく下げる案 — 観測は 2 rune 差のみ。
