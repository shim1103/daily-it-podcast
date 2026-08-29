---
name: TextWriter brief Prompt は 1 本化し field ごとに guideline と limits placeholder を結合する
date: 2026-08-29T17:00:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. TextWriter brief の固定 Prompt は **`entities/constants/text_writer_brief_prompt.go` の `TextWriterBriefPrompt` 1 本**を正とする。`PodcastBriefSettings` のような分割 file は置かない。口調・番組設定は template 先頭の日本語 prose に inline する。
2. **`# Character limits` と `# Field guidelines` を分けない**。field ごとに `# title` / `# intro` / `# topics` / `# topic title` / `# topic preface` / `# topic detail` / `# closingSummary` / `# total length` の section を設け、各 section 内で guideline と文字数 placeholder を並べる。
3. 数値 placeholder（`{{TITLE_MIN}}` 等）は **`manuscript_draft_limits.go` が SSOT**。`ComposeBrief` は文案を組み立てず、`embedManuscriptDraftLimits` で定数を埋め込むだけとする。
4. **動的埋め込み**は `{{SOURCES}}` と `{{JSON_EXAMPLE}}` の 2 種のみ。`ComposeBrief` が SourceItem 列と WriterOutput example を生成する。
5. `# {section-header}` 行のみ **English**。本文・箇条書きは **日本語**。

## 2. Reason

1. 口調設定と field 制約は同じ brief の一部であり、file 分割は shim の編集単位を不必要に割る。
2. guideline と limits を field 単位で並べると、Cursor が field ごとの制約を読み取りやすい。
3. `{{LIMITS}}` ブロックを別組み立てすると「誰が文案を書くか」が曖昧になる。template に placeholder を置き build が定数注入するだけにすると SSOT が limits に一本化される（16-42 の Rejected 案 3 と整合）。

## 3. Rejected

1. `PodcastBriefSettings` と `TextWriterBriefPrompt` の 2 file 分割 — 編集単位の過剰分割。本 Decision で 1 本化。
2. `{{LIMITS}}` を ComposeBrief が prose 組み立てする案 — build が Domain 文案を増やす。placeholder 注入に限定する。
3. closingSummary 専用の min/target/max 定数を Decision 時点で新設する案 — parse 定数が未存在。guideline のみ記載し、数値は将来 limits 追加時に placeholder を足す。
