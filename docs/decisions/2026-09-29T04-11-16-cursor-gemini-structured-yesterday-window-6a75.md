---
name: WriterOutput schema の正本は entities/models
date: 2026-09-29T04:11:16
branch: cursor/gemini-structured-yesterday-window-6a75
supersedes: 2026-09-29T02-10-39 §Decision.2（contracts 配置）
---

## 1. Decision

1. WriterOutput の JSON Schema 正本は `apps/generator/internal/entities/models/writer_output.schema.json`（`models.WriterOutputSchema`）。
2. `geminiapi` は `models.WriterOutputSchema` を `responseSchema` に載せる。`contracts/` に WriterOutput schema を置かない。
3. brief の `{{JSON_EXAMPLE}}` fixture も同 dir の `models.WriterOutputExampleJSON`。`application/build` はそれを読むだけ。
4. `contracts/` は完成 manuscript など generator / playback 共有 wire だけを持つ。

## 2. Reason

1. `WriterOutput` の `json.Unmarshal` 正本は既に `entities/models`。schema を別 module（`contracts/`）へ出すと、wire 形の所有者が二つに割れる。
2. WriterOutput は generator 途中型であり playback は読まない。共有 `contracts/` へ載せる必然が無い。
3. Gemini Adapter が読む schema と Domain の Unmarshal 型を同 dir に置くと、形の変更が 1 箇所に寄る。

## 3. Rejected

1. **`contracts/writer-output.schema.json` を正本にする案**（旧 Decision）— playback 非依存の途中型を共有契約へ載せ、Unmarshal 正本と schema が別 module になる。
2. **geminiapi 内に schema JSON を直書きする案** — WriterOutput 形の第二 SSOT。
3. **example だけ application/build に残す案** — schema と example が分かれる。両方とも wire 例示なので models に揃える。
