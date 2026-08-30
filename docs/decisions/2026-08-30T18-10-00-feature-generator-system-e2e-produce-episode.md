---
name: TextWriter draft 検証失敗は有限再試行する
date: 2026-08-30T18:10:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `ProduceEpisode` は `TextWriter.Write` → `ManuscriptDraftFromWriterOutput` を最大 `TextWriterMaxAttempts`（3）回繰り返す。
2. draft 検証が成功した時点で打ち切る。`Write` 自体の error は再試行せず即 return する。
3. 上限まで失敗したら最後の `invalid_manuscript_draft` を返す。

## 2. Reason

1. System 実測（run 33303058840）で JSON 受理後に `topic[2].preface rune count 133 is out of range [140, 252]` となり、LLM の長さ揺れが観測された。
2. Domain Rule（rune range）は緩めず、再生成で吸収する方が契約を保つ。
3. Gemini TTS と同様に有限上限を置き、無限 retry を防ぐ。

## 3. Rejected

1. Domain 下限を下げる案 — 朗読尺の正本を崩す。
2. 短すぎる field を機械 padding する案 — 原稿意味を変える。
3. Write の Infrastructure Error も再試行する案 — Cursor API 到達失敗は別原因。今回の観測範囲外。
