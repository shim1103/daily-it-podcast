---
name: TextWriter brief prompt と JSON_EXAMPLE を更新する
date: 2026-09-14T18:54:42
session_id: none
branch: feature/generator-textwriter-brief-prompt-update
prev: なし
---

## 1. Summary

TextWriter brief prompt に Purpose 選出 Workflow・target 狙い・draft meta 禁止・聞きやすい語り契約を載せ、`{{JSON_EXAMPLE}}` を git 管理の検証済み WriterOutput JSON embed へ切り替えた。generator-lane の P1/P2 行を実装済みへ更新し、残作業を draft-rate 再測へ寄せた。

## 2. Changes

1. brief template を Purpose1/2 半々選出・link 必見・target 付近・title=main・intro 列挙可へ更新
2. `writer_output_example.json` を embed し `ManuscriptDraftFromWriterOutput` で検査してから brief へ埋める
3. prompt test は数値 / 動的 placeholder のみに縮小。example 読込と ComposeBrief 経由の valid 固定を build package へ追加
4. `origin/develop` を merge して先端へ同期

### Commits

- `b284e49`
- `09b06fc`
- `c7db349`
- `520adff`
