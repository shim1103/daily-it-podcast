---
name: playback web↔worker HTTP 契約と todo 切り出しの完了
date: 2026-08-17T19:10:00
session_id: none
branch: feature/playback-web
---

## 1. Summary

web↔worker の **playback HTTP 契約**（List/Get の TS schema、HTTP status→契約 `code` の分類）を `apps/playback/contracts/` に固定し、`http.test.ts` / `http-error.test.ts` を追加して unit で green にした。あわせて `contracts/drive-layout.md` の一覧読取規則を「一覧は `*.json`」へ更新し、`DESIGN.md` / `README.md` の置き場を整えた。さらに worker 中心の開発を進めるため、`docs/tasks/todo/` に **3つの分割 task**（worker episodes / worker http / web api-client）を追加し、`playback-lane.md` に依存関係を追記した。

## 2. Changes

1. playback HTTP 契約 `apps/playback/contracts/` を追加（List/Get schema、`audioRef`、path 定数・関数）
2. HTTP status 分類 `classifyHttpStatus` を unit で検証
3. `zod` を playback に追加し、境界 schema の runtime parse を可能に
4. `contracts/drive-layout.md` を一覧 `*.json` 方針へ修正
5. `DESIGN.md` / `README.md` の参照元を整理し、HTTP 契約の置き場を明示
6. `docs/tasks/todo/` に 3 task を追加し、`playback-lane.md` に依存を追記

### Commits

- `3e710a3` — feat(playback): web↔worker HTTP契約（List/Get）と契約テスト
- `53e7bff` — docs(playback): web↔worker HTTP契約のtodo切り出しと運用整理
