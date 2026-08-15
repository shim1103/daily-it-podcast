---
name: X 取得の境界 stub と Issue 下書き
date: 2026-08-15T19:00:09
session_id: none
branch: feature/x-api-adoption
prev: なし
---

## 1. Summary

非公式 X API の採用・後回し範囲を decision に残し、Generator に `PostSource` / `Post` / 監視定数の境界 stub を追加した。Issue1・2 の draft を tasks に置き、Issue3（GetXAPI）は generator-lane に未完了 task として追記した。

## 2. Changes

- X API 採用 decision と後回し decision を追加
- `PostSource` Port、`Post` / `Media`、`WatchUserIDs` / `FetchWindow`、`go.mod` を追加
- DESIGN / README に試作秘密名を追記
- Issue1・2 draft md を作成、Issue3 詳細を generator-lane に追記

### Commits

- `docs(decisions): X API 採用と後回し範囲を記録する` — 採用と後回しの SSOT
- `feat(generator): PostSource 境界 stub を追加する` — Port / Domain / 定数
- `docs: 情報取得の秘密名と Issue 下書きを残す` — README・DESIGN・tasks
