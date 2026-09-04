---
name: TextWriter 再試行を 5 回にし correction 文言を topic 数にも効かせる
date: 2026-08-30T22:15:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `TextWriterMaxAttempts` を 3 から 5 へ上げる。
2. 再試行 brief の correction は「文字数」限定ではなく、topic 件数を含む検証失敗全般の解消を指示する。

## 2. Reason

1. run 33313398040: 3 回試しても `topic count 2 is out of range [3, 7]`。旧 correction が「文字数を直し」だけだと件数不足の誘導が弱い。
2. Gemini timeout は別 Decision で延ば済み。今回の赤は draft 検証側。

## 3. Rejected

1. `DraftTopicCountMin` を 2 へ下げる案 — 製品尺の意図変更。再試行強化で足りるか先に見る。
2. Source 件数 < 3 で即 Domain Error にする案 — 別 Decision。今回の切り分けは再試行文言と回数。
