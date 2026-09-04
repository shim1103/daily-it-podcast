---
name: TextWriter draft 検証失敗は有限再試行する
date: 2026-08-30T18:10:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `ProduceEpisode` は `TextWriter.Write` → `ManuscriptDraftFromWriterOutput` を最大 `TextWriterMaxAttempts`（3）回繰り返す。
2. draft 検証が成功した時点で打ち切る。`Write` 自体の error は再試行せず即 return する。
3. 2 回目以降の brief には、元 brief に加えて前回の draft 検証 error 本文を `# Previous attempt rejected` 節として付ける。
4. 上限まで失敗したら最後の `invalid_manuscript_draft` を返す。

## 2. Reason

1. System 実測（run 33303058840 / 33305956174）で JSON 受理後に preface rune 数が下限未満となり、盲目再試行だけでは同じ短さが出た。
2. Domain Rule（rune range）は緩めず、失敗理由をモデルへ返す方が次案の修正確率を上げる。
3. Gemini TTS と同様に有限上限を置き、無限 retry を防ぐ。

## 3. Rejected

1. Domain 下限を下げる案 — 朗読尺の正本を崩す。
2. 短すぎる field を機械 padding する案 — 原稿意味を変える。
3. Write の Infrastructure Error も再試行する案 — Cursor API 到達失敗は別原因。今回の観測範囲外。
4. 同じ brief の盲目再試行のみ — run 33305956174 で 3 回とも短 preface のまま失敗した。
