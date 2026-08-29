---
name: TextWriter 出力の parse と Domain validation は entities 定数と ManuscriptDraft 型が正とする
date: 2026-08-29T14:11:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. Cursor 返却 `string` → `ManuscriptDraft` の解釈は **`manuscriptDraftFromWriterOutput`（Builder）** が行う。失敗は Domain Error（`invalid_manuscript_draft`）。Infra Error にしない。
2. Domain Rule の正本は **`entities/constants` の manuscript / writer-output 定数**と **`entities/models.ManuscriptDraft`（`Title` 含む）** である。prompt に書く制約のうち機械判定可能なものは **すべて parse で検証**する。
3. 各朗読 field（`Title` / `Intro` / topic の `Title` / `Preface` / `Detail` / `ClosingSummary`）は **日本語を 1 文字以上含み**、trim 後非空、**末尾が `。`** であること。無意味な空白・改行のみの field は reject する。
4. field 文字数 range・topic 数 range・**全体文字数 range** は定数どおり parse する。prompt 用の推奨平均は定数に載せてよいが **parse 対象にしない**。
5. `WriteEpisode`（Gate）は **Contract validation のみ**（`manuscript.schema.json` 等）。文字数 range や日本語含有は Gate の責務にしない。

## 2. Reason

1. 生成方針（何を valid とするか）と永続前検査（schema）は Builder / Gate で既に分離されている。内容品質を Gate に寄せると SRP が崩れ、lesson 109 の検証ゲートと混ざる。
2. prompt 约束と parse 规则を定数 SSOT に揃えると、Cursor 出力の drift を Domain Error で観測でき、Infra 障害と区別できる。
3. 全体文字数 range は TTS 入力総量と WAV 尺に直結する。field 単位だけでは episode 全体の上限を抑えられない。

## 3. Rejected

1. 文字数・句点规则を `WriteEpisode` でも二重検査する案 — Gate が Domain 方針を知る。schema `minLength` だけで足りる。
2. parse 规则を Decision 本文に数値一覧として再掲する案 — 正本は constants。Decision は参照のみ（DRY）。
3. 推奨平均も parse する案 — 平均は prompt 誘導用。fail 条件にすると Cursor の自然な揺れで false negative が増える。
