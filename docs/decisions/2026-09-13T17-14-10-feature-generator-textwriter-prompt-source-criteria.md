---
name: 各情報源の raw→SourceItem 写像は Adapter 契約として固定し、裁量に残さない
date: 2026-09-13T17:14:10
branch: feature/generator-textwriter-prompt-source-criteria
---

## 1. Decision

1. **各情報源 Adapter の raw→`SourceItem` 写像は固定する。** 実装者の裁量に「どの raw を Summary / Detail / Discourse / Meta のどこへ載せるか」を残さない。
2. **字段形・URL の Purpose 側・空の意味**は Decision `2026-09-13T17-14-00` を正とする。本 Decision は「写像を固定する」方針だけを持つ。
3. **写像表（raw 名と行き先の一覧）の正本は Decision に置かない。**
   - 実装済み Adapter（現状 HackerNews / Lobsters）→ 各 package の `toSourceItem`（A）
   - 未実装 Adapter（Publickey / TechCrunch / CloudWatch の `List`）→ C の Issue file の Acceptance、実装後は同様に Adapter code が正本
4. engagement（score 等）を載せない（先行 `2026-09-02T14-41-02`）。
5. 本 Decision は session Decision `2026-09-13T15-09-10` の「写像は未確定のまま空」を **supersede** する。

## 2. Reason

Purpose 空判定（`Detail` / `Discourse` の Text+Links）は URL の置き場に依存する。源ごとに「外部記事 URL か議論ページか」が違うため、写像を裁量にすると同じ問いが Adapter ごとに再発する。

一方、raw 字段名と行き先の対応表は **境界契約・達成契約の百科**であり、Decision の正本にすると logging（契約値の再掲禁止）と scope-split B（理由・Rejected のみ）に反する。方針（固定する／どこが正本か）だけを Decision に残し、表は A または C へ置く。

## 3. Rejected

1. **写像を Adapter 実装者の裁量に残す案** — Purpose 空判定と URL 置き場が源ごとにぶれる。
2. **raw→field 対応表を Decision 本文に置く案** — 仕様・契約の正本を Decision に二重化する。正本は A（実装後）/ C（実装前）。
3. **HN/LOB の議論 permalink を `Detail.Links` に置く案** — 議論ページは P2（`17-14-00`）。
4. **報道源の記事 URL を `Meta` に置く案** — `17-14-00` Rejected と同型。
5. **engagement（score 等）を載せる案** — `14-41-02` Rejected を維持。
