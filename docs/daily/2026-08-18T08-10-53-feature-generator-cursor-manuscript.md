---
name: Cursor TextWriter Port と decision/task draft
date: 2026-08-18T08-10-53
session_id: none
branch: feature/generator-cursor-manuscript
prev: none
---

## 1. Summary

Generator に **Cursor CLI** の呼び出しを driving する `TextWriter` Port と argv 定数を追加し、Cursor envelope 解釈・呼び出し規則の Decision および Adapter 実装 Issue draft（todo）を置いた。

## 2. Changes

1. `TextWriter` Port（断片生成）を generator application に追加した
2. Cursor CLI argv を決める定数（`composer-2.5` / `ask` / `--output-format json` / `--sandbox enabled` / `--trust`）と `CURSOR_API_KEY` という秘密名を generator に追加した
3. `DESIGN.md` と `docs/tasks/todo/generator-lane.md` を更新して、この task の配置先を固定した
4. `docs/decisions/2026-08-18T16-30-00-feature-cursor-text-writer.md` と `docs/tasks/todo/generator-cursor-text-writer.md` を追加した

### Commits

- `5a55471` — feat(generator): add Cursor TextWriter Port and argv constants
- `a0794e6` — docs(generator): add Cursor TextWriter decision and task draft

