---
name: docs status audit と playback web 層の土台
date: 2026-08-20T19:37:00
session_id: 52aac738-b739-4cb7-9aca-4e63ef185228
branch: docs-status-audit
prev: なし
---

## 1. Summary

lane チェックリストと DESIGN を merge 済み実装へ同期し、playback web の role ↔ dir 対応を decision へ固定した。`view-models/` と `lib/` の空 dir と Vite・Pico.css の package 明示宣言で、後続の toolchain / UI Issue が同じ前提で着手できる土台を残した。

## 2. Changes

1. `generator-lane.md` と `playback-lane.md` の Port 名・完了項目・Issue 化待ち表を code と PR 履歴へ合わせた。
2. `DESIGN.md` の `ItemSource` 表記と web 層 dir 一覧を実装・decision へ合わせた。
3. `playback-lane.md` の依存図を訂正し、`web-api-client` と `toolchain` が並行できる形へ直した（旧図の直列は誤り）。
4. scope-split の C（toolchain / UI の Issue draft・worktree）は意図的に未実施。誤って先取りした分は session 内で一度 revert 済み。
5. 検証：`typecheck` / `lint` / `format:check` / `test:unit` を実行し 118 test 通過。

### Commits

- `45cce82`
- `3803cd9`
- `adba7a0`
- `b6d8b1f`
- `3bf5315`
- `224d1c0`
