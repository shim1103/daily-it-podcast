---
name: generator 情報取得 Port を ItemSource へ移行
date: 2026-08-19T14:29:20
session_id: none
branch: refactor/generator-source-port
prev: なし
---

## 1. Summary

Generator の情報取得を `PostSource` / `Post` から `ItemSource` / `SourceItem`（必須 `SourceID` + `OccurredAt`、残り `Context`）へ移した。Application は `FetchSourceItems` が `List` 1回のみ呼ぶ。X 両 Adapter は空 `List` の Stub とし、旧 `ListByUser` 実装は各 `post_source.go` の `todo:` block comment に参照として残した。decision・todo・lane を整備し、4 commit に分割して push した。

## 2. Changes

1. `docs/decisions/2026-08-19T13-25-20-refactor-generator-source-port.md` を追加
2. `ItemSource` Port / `SourceItem` / `FetchSourceItems` と sociable unit を追加。`Post` / `PostSource` / `FetchWatchedPosts` を削除
3. GetXAPI / TwitterAPI.io を `ItemSource.List` Stub へ載せ替え。旧 vendor 取得 code は block comment 参照
4. `docs/tasks/todo/generator-x-item-source.md` を追加（GitHub Issue 化しない）。`generator-lane.md` を更新
5. `docs/lessons/index.md` に 82–85 を追記
6. `./scripts/generator/test-unit.sh` と pre-commit（generator static + playback biome/tsc + unit）が pass

### Commits

1. `2d2a257`
2. `9d51b7a`
3. `dab7095`
4. `66b4631`
