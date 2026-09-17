---
name: runtime図とDESIGN/DEPLOY/laneをfallback実装へ追従
date: 2026-09-17T14:06:52
session_id: none
branch: feature/generator-workflow-fallback-wiring
prev: 2026-09-17T10-55-31
---

## 1. Summary

4つのstacked PR（credential decision・TTS fallback・TextWriter fallback・workflow yml統一）で実装した内容が、code-first runtime構成図（`apps/diagrams/runtime.py`）と設計SSoT文書（DESIGN.md・DEPLOY.md）に反映されていなかったため追従させた。あわせて`docs/tasks/todo/generator-lane.md`の「済み」セクションを、run ID・SU/Narrow内訳などの進捗ログを削って要約へ圧縮した。

## 2. Changes

1. `apps/diagrams/runtime.py`の原稿・TTS経路を、旧設計（Cursor primary→Gemini fallback 1段、TTS fallbackなし）から現行実装（Gemini free→Cursor→Gemini paid finalの3段、TTS Gemini free→Gemini paid finalの2段）へ更新し、PNGを再生成した
2. `DESIGN.md`§3外部I/O表の原稿・TTS行を、fallback順序を示す記述へ更新した
3. `DEPLOY.md`のcredential一覧・`generator-draft-rate.yml`/`generator-system.yml`の使う登録・rate計測workflow記述にあった`TEST_CURSOR_API_KEY`/`TEST_SPARE_GEMINI_API_KEY`等の旧登録名を、Decision 2026-09-16T00-39-21準拠の現行登録名へ揃えた
4. `docs/tasks/todo/generator-lane.md`の「済み」1〜16番（storage移行・fallback初期実装等の詳細な進捗ログ）を7項目の要約へ圧縮し、D（未決）セクションから今回の4PRで解決済みになった項目を削除した

### Commits

- `1ecef49`
- `515671f`
