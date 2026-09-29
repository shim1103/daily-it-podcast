---
name: WriterOutput の responseSchema は型から生成する
date: 2026-09-29T04:53:13
branch: cursor/gemini-structured-yesterday-window-6a75
supersedes: 2026-09-29T02-10-39 §Decision.2（schema 置き場）; 2026-09-29T04-11-16（手書き schema file）
---

## 1. Decision

1. Gemini `responseSchema` は **`models.WriterOutput` の json tag から生成**する（`models.WriterOutputSchema`）。手書き schema file も `contracts/` 配置も採らない。
2. wire Unmarshal 正本は `WriterOutput`。`ManuscriptDraft` に json tag は付けない。
3. schema が保証するのは wire の **形**まで。件数・rune・日本語・句点は `ManuscriptDraftFromWriterOutput`（buildFn）が正本。
4. wire / draft の bookend field は **`openingIntro` / `endingSummary`**（完成稿の opening / ending と対称）。

## 2. Reason

1. 手書き schema や共有 `contracts/` へ途中型を置くと、Unmarshal 型と形の所有者が割れる。tag 生成なら 1 箇所。
2. Domain 検証は既に Application の buildFn。schema は Gemini への形ヒントだけ。
3. 無修飾 `summary` は SourceItem や番組要約とぶつかる。`endingSummary` で bookend を示す。

## 3. Rejected

1. **`contracts/writer-output.schema.json` を正本にする案** — playback 非依存の途中型を共有契約へ載せる。
2. **手書き `.schema.json` を models 隣に置く案** — 型と第二 SSOT。
3. **geminiapi 内へ schema JSON を直書きする案** — Adapter 固有の第二 SSOT。
4. **`ManuscriptDraft` に json tag を付けて wire 正本にする案** — Domain 途中型と wire が混ざる。
5. **field を無修飾 `summary` にする案** — 他 `Summary` と衝突する。
