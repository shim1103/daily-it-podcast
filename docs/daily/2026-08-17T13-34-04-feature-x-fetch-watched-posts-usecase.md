---
name: 監視 user 一括取得 UseCase 追加
date: 2026-08-17T13:34:04
session_id: unknown
branch: feature/x-fetch-watched-posts-usecase
prev: 2026-08-17T13-12-00-feature-x-post-source-adapter.md
---

## 1. Summary

監視 user 一括取得 UseCase（`FetchWatchedPosts`）を Application 層に追加し、Fake `PostSource` 付き sociable unit test を通した。完了した Issue2 draft todo を削除し、generator-lane を更新した。

## 2. Changes

- `application.FetchWatchedPosts` / `NewFetchWatchedPosts` / `Run` を追加（`WatchUserIDs` × `FetchWindow`、fail-fast）
- Fake `PostSource` と sociable unit test を追加
- `docs/tasks/todo/x-fetch-watched-posts-usecase.md` を削除
- `docs/tasks/todo/generator-lane.md` の当該項目を完了表示に更新

### Commits

- 未 commit
