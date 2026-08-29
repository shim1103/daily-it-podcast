---
name: ProduceEpisode の TTS 単位は Greeting / Intro / Preface / Detail / Summary / Farewell を各 1 Synthesize とする
date: 2026-08-29T14:10:00
branch: docs/produce-episode-run-spec
supersedes_in_part: 2026-08-25T22-37-29-feature-generator-cmd-usecase-boundary.md
---

## 1. Decision

1. `ProduceEpisode`（Builder）の TTS 順序は、各 topic について **Preface と Detail を別々**に `SpeechSynthesizer.Synthesize` する。Opening は **OpeningGreeting（date 注入済み）と Intro を別々**、Closing は **ClosingSummary と ClosingFarewell を別々**とする。
2. `manuscript.schema.json` の `body.topics[].startSec` は **当該 topic の Preface 朗読開始時刻**とする。Detail 尺は topic 内に含めるが startSec には含めない。
3. `port.SpeechSynthesizer` の署名は **1 text → 1 WAV のまま**変更しない。segment 間の pause は Port 引数ではなく、分割呼び出しと `concatWAV`（別 Decision）で表現する。
4. 本 Decision は `2026-08-25T22-37-29` のうち「opening = 定型+Intro を 1 speakable」「topic = preface+detail を 1 speakable」の結論を上書きする。Builder が TTS 順序を所有する側面は維持する。

## 2. Reason

1. segment 間に無音を挟む要件は、1 本の speakable 文字列結合では制御できない。Preface / Detail を分ければ、API 呼び出し境界と pause 境界を一致させられる（Orthogonality）。
2. `startSec` は UI の topic シーク用であり、topic ブロックの先頭（Preface）で十分である。Detail を同じ startSec に含めるのは schema 上も自然である。
3. Port に segment 配列や pause 秒を載せると、Application が vendor 演出能力の形を知る。既存 Decision（SpeechSynthesizer 薄さ）と衝突する。

## 3. Rejected

1. Preface+Detail・Greeting+Intro を各 1 `Synthesize` にまとめる案 — pause を文字列メタ指示で表現する必要が出る。TTS Adapter の envelope 外に演出文を書くと朗読されるリスクがある。
2. `SpeechSynthesizer` Port に pause / voice / 口調引数を足す案 — vendor 追随のたび Port が変わり、Application の vendor 非依存が崩れる（別 Decision で Port 安定を固定）。
3. 固定 Greeting / Farewell を Drive 上の WAV から読む案 — date を挨拶に埋め込むため、Builder が TTS 前に文案を組み立てる必要がある（別 Decision）。
