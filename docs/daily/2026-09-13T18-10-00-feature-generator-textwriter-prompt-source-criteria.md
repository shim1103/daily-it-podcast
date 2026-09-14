---
name: 情報源再編と SourceItem SourceBody 化の session
date: 2026-09-13T18:10:00
session_id: ad475b0f-39a0-40cc-9db9-cf70153c219e
branch: feature/generator-textwriter-prompt-source-criteria
prev: なし
---

## 1. Summary

ITmedia を外し Publickey / TechCrunch / CloudWatch を結線する方針と、`SourceItem` の Summary / Detail / Discourse（Text+Links）/ Meta 契約を固定した。写像表は Decision から外して C Issue へ移した。新 3 源の `List` 実装は未着手（stub）。

## 2. Changes

1. Decision 4 本（源セット・旧 string 形・SourceBody・写像固定方針）と地図参照更新
2. Generator 契約変更と HN/LOB 写像、新 3 Adapter stub、ITmedia 削除
3. C Issue（create-issue template）と generator-lane 更新
4. `lessons/index.md` は `origin/develop` を正本に取り、本 session 知見を末尾追記（branch 全体の rebase はしない）

### Commits

- `6c30557`
- `4b57620`
- `d8c32e9`
