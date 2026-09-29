---
name: WriterOutput schema は型から生成し wire field は openingIntro / endingSummary
date: 2026-09-29T04:53:13
branch: cursor/gemini-structured-yesterday-window-6a75
supersedes: 2026-09-29T04-11-16（手書き writer_output.schema.json 正本）
---

## 1. Decision

1. Gemini `responseSchema` は **`models.WriterOutput` の json tag から生成**する（`models.WriterOutputSchema`）。手書き `writer_output.schema.json` は置かない。
2. wire / Domain draft の対称 field 名は **`openingIntro` / `endingSummary`**（Go: `OpeningIntro` / `EndingSummary`）。旧 `intro` / `closingSummary` は使わない。
3. `ManuscriptDraft` に json tag は付けない。wire Unmarshal 正本は `WriterOutput` のまま。
4. Domain 制約（件数・rune・日本語）は schema に載せない。`ManuscriptDraftFromWriterOutput`（buildFn）が正本。

## 2. Reason

1. 手書き schema JSON は `WriterOutput` と第二の SSOT になる。tag から生成すれば形の変更は型 1 箇所。
2. `opening` / `ending` bookend と揃えると intro/closing 接頭辞の非対称が消える。
3. Domain 検証は既に Application の buildFn。schema は Gemini への形ヒントだけ。

## 3. Rejected

1. **手書き `.schema.json` を models 隣に置く案**（旧）— Unmarshal 型と二重管理。
2. **`ManuscriptDraft` に json tag を付けて wire 正本にする案** — Domain 途中型と wire が混ざる（既存 WriterOutput Decision と衝突）。
3. **field を無修飾 `summary` にする案** — `SourceItem.Summary` や番組要約とぶつかる。`endingSummary` で bookend を示す。
