---
name: Cursor CLI を Cloud Agents REST stub へ移し pr-completion
date: 2026-09-03T17:46:27
session_id: none
branch: feature-generator-cursor-cli-to-http-api
prev: なし
---

## 1. Summary

原稿 TextWriter の transport を Cursor CLI / commandlaunch から Cloud Agents REST へ移す契約（A stub）と Decision（B）を固定し、本実装用の C task 1 本と地図 docs を現行前提へ揃えた。Write 本実装は未着手。

## 2. Changes

1. Explore で Cloud Agents REST が chat-completions ではなく agent workflow であること、no-repo / SSE / retry / Pro 制約を整理した。
2. scope-split 分類後、B Decision を main が書き、executor に A（stub・削除・CI）を委譲。空 dir 残存と未使用 `infraErr` を manager 監査で修正させた。
3. `sharedCursorHTTPClient` を vendor 非依存の `sharedHTTPClientWithoutTimeout` へ改名。C を `generator-cursor-http-text-writer.md` 1 本に束ね、DESIGN / README / lane を更新した。
4. scope-split / decisions skill に「plan 時 A/B file-crud-plan 必須」「1 question ≠ 1 file」を追記した（dotfiles skills）。

### Commits

- `3f01a45`
- `6cfab90`
- `6b8225f`
