---
name: 監視 user 一括取得 UseCase 追加
date: 2026-08-17T13:34:04
session_id: unknown
branch: feature/x-fetch-watched-posts-usecase
prev: 2026-08-17T13-12-00-feature-x-post-source-adapter.md
---

## 1. Summary

監視 user 一括取得 UseCase（`FetchWatchedPosts`）を Application 層に追加し、Fake `PostSource` 付き sociable unit test を通した。Issue2 draft todo を削除し generator-lane を更新した。skill 準拠の reviewer 査読と GWT 復元を経て commit・push した。PR #14 を作成し、integration CI は SUCCESS。Copilot は quota 超過で review 不可。`origin/develop` との conflict は無し。

## 2. Changes

- `application.FetchWatchedPosts` / `NewFetchWatchedPosts` / `Run` を追加（`WatchUserIDs` × `FetchWindow`、fail-fast）
- Fake `PostSource` と sociable unit test を追加（先頭失敗・後続失敗の部分結果破棄を含む）
- test に GWT 構造 label を付与
- `docs/tasks/todo/x-fetch-watched-posts-usecase.md` を削除
- `docs/tasks/todo/generator-lane.md` の当該項目を完了表示に更新
- PR #14（base: develop）。`shim gh create-pr` は `workflow/project.toml` 欠如で失敗したため `gh pr create` で代替

### Commits

- `e97e7b4` — feat(generator): 監視 user 一括取得 UseCase を追加する
- `37ee062` — docs: UseCase todo を片付け generator-lane を更新する
- `f6ac8d7` — docs(log): セッションログ
