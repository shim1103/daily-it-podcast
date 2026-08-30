---
name: brief 組み立ては application/build.ComposeBrief が所有し見出し定数 file は置かない
date: 2026-08-29T16:30:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. TextWriter へ渡す **brief** の組み立ては **`application/build.ComposeBrief(items []SourceItem) string`** が所有する。`ProduceEpisode.Run` は fetch 結果を渡して呼ぶだけとする。
2. brief の **固定散文**（一人喋り口調・人物・番組設定）は **`entities/constants/podcast_brief_settings.go` の 1 多行 const** に置く。文言の正本はここ（shim が編集）。parse しない。
3. brief の **骨格 template**（英語 `#` 見出し・出力指示・field ニュアンス文）は **`application/build` 内の template 文字列** に置く。`entities/constants/brief_sections.go` のような **見出し定数 file は置かない**。
4. brief 内の **数値 range** は `entities/constants/manuscript_draft_limits.go` から **ComposeBrief が fmt 注入**する（parse する定数のみ SSOT 二重化を避ける）。
5. 出力 JSON の field 名・形の指示は **`entities/models.WriterOutput` を正**とし、example は ComposeBrief が models から生成する。挨拶（OpeningGreeting / ClosingFarewell）は brief に含めない。
6. ソース節の見出しは **`# Source`** とする。窓幅（24 時間等）の説明文は brief に書かない。各 item の時刻は **`OccurredAt` を brief 内テキストとして載せる**だけで足りる。
7. brief 見出しの `#` 行は **English** とする（例: `# Source`, `# Output`, `# Character limits`）。

## 2. Reason

1. TextWriter Port は `brief string` のみ。prompt 方針を Adapter や Composition に閉じると Builder 境界が分散する（既 Decision `14-17` の Rejected と同型）。
2. parse しない見出しを constants に切ると、template 変更のたびに file が増え、SSOT が不明瞭になる（KISS / YAGNI）。template 1 本 + podcast 設定 1 const で足りる。
3. `models` は wire / Domain 型の置き場である。prompt 散文を models に入れると Entities が LLM 指示文を知る（層汚染）。
4. 窓幅は `FetchWindow` と Fetch が既に決める。brief に「24 時間」を書くと Fetch 方針変更時に二重管理になる（DRY）。

## 3. Rejected

1. `entities/constants/brief_sections.go` に見出し定数を置く案 — parse しない文字列の過剰定数化。`14-17` 当時の置き場を本 Decision で置き換える（旧 file は A から削除）。
2. brief 組み立てを Cursor Adapter 内に置く案 — Infra が episode 生成方針を知る。
3. podcast 設定散文を Decision 本文に再掲する案 — 文言 SSOT は constants。Decision は参照のみ。
4. `# SourceItems` や日本語見出し `# 出力形式` を正とする案 — 本 Decision で English `#` に統一。
5. brief に「過去 24 時間」等の窓説明を書く案 — Fetch が窓を所有。ソース行の `OccurredAt` で足りる。
