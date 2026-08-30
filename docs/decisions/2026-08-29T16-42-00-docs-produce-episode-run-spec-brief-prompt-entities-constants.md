---
name: TextWriter brief の固定 Prompt は entities/constants に置き build は埋め込み組み立てのみ行う
date: 2026-08-29T16:42:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. TextWriter brief の **固定 Prompt 文言**（英語 `#` 見出し + 日本語本文 + プレースホルダ）は **`entities/constants`** に置く。`application/build` に Prompt 散文を hardcode しない。
2. `entities/constants` の Prompt 系は **parse しない embed 専用**とする。`PodcastBriefSettings`（口調・番組）と `TextWriterBriefPrompt`（template + `{{PODCAST_SETTINGS}}` / `{{SOURCES}}` / `{{LIMITS}}` / `{{JSON_EXAMPLE}}`）の 2 本を正とする。
3. **parse / validation 用の数値定数**は **`manuscript_draft_limits.go` のみ**が正とする。ComposeBrief が `{{LIMITS}}` をここから組み立てて埋め込む。
4. **`application/build.ComposeBrief`** は **Domain 固有の埋め込み組み立て**のみ行う。`{{SOURCES}}`（SourceItem 列）、`{{LIMITS}}`（limits 文案）、`{{JSON_EXAMPLE}}`（WriterOutput 形）を生成し、constants の template へ差し込む。Prompt 本文そのものは変更しない。
5. `# {section-header}` 行のみ **English**。見出し下の指示文・箇条書きは **日本語**。

## 2. Reason

1. episode 生成の **Domain 方針文**は Entities 層の定数に置くのが、この repo の Gate/Builder 分割と整合する。Application/build は orchestration と動的埋め込みに専念する（SRP）。
2. Prompt を build に置くと、Domain 文案変更が Application 詳細 package を触ることになり、層の読み取りが崩れる。
3. parse する定数と embed だけの Prompt を同 dir で **file 責務で分ける**と、shim が「数値だけ変える」「文案だけ変える」を迷わない（Least Astonishment）。
4. 見出しだけ English にすると、Cursor への section 認識と日本語出力指示を両立できる。

## 3. Rejected

1. Prompt 骨格を `application/build/brief_template.go` に置く案 — Domain 文案が Application/build に漏れる。本 Decision で置き換える（旧 file は A から削除）。
2. 見出しも含め brief 全文を日本語にする案 — section 認識のため `#` 行は English に統一。
3. `{{LIMITS}}` 文案を Prompt 定数に hardcode し数値だけ差し替える案 — 数値 SSOT が limits と Prompt に分裂する（DRY 割れ）。
