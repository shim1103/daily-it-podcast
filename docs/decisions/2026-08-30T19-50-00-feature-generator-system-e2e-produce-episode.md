---
name: topic.preface 下限を 10 秒へ下げる
date: 2026-08-30T19:50:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `DraftTopicPrefaceMinSec` を 20 から 10 へ下げる（文字数は `CharsPerSecond` 畳み込みで 70）。
2. Target / Max（28 / 36 秒）は変えない。

## 2. Reason

1. Prompt は「短い前置き」と書いているのに旧下限 20 秒は短くない。文言と定数が矛盾していた。
2. System 実測で preface が 99 rune（run 33305956174）・133 rune（run 33303058840）と繰り返し下限 140 未満になり、有限再試行でも同じ傾向が残った。
3. 下限を 10 秒にすれば観測値を Domain Rule 内に収めつつ「短い」意図と揃う。

## 3. Rejected

1. 再試行回数だけ増やす案 — モデルが系統的に短いため回数では収束しない。
2. preface を機械 padding する案 — 原稿意味を変える。
3. Target / Max も下げる案 — 今回の失敗は下限超過のみ。全体尺 target 整合を壊さない。
