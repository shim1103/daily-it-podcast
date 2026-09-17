---
name: runtime図とDESIGN/DEPLOY/laneをfallback実装へ追従
date: 2026-09-17T14:06:52
session_id: none
branch: feature/generator-workflow-fallback-wiring
prev: 2026-09-17T10-55-31
---

## 1. Summary

4つのstacked PR（credential decision・TTS fallback・TextWriter fallback・workflow yml統一）で実装した内容が、code-first runtime構成図（`apps/diagrams/runtime.py`）と設計SSoT文書（DESIGN.md・DEPLOY.md）に反映されていなかったため追従させた。あわせて`docs/tasks/todo/generator-lane.md`の「済み」セクションを、run ID・SU/Narrow内訳などの進捗ログを削って要約へ圧縮した。

T2/T3の実装コードが参照していたが、どちらのPRにも含まれていなかったdecision文書2件（`sources []port.TextWriter`配列化・`ErrDraftRejected`分離、`LastAttempt`部分情報持ち越し・`wrapIfSourceExhausted`廃止）を本PRへ補完した。

master push契機のCD連鎖を、`generator-system.yml`（master向けpull_request→push）を起点に、成功時だけ`playback-deploy.yml`（`apps/playback/**` path絞り込み→`generator-system`成功時のworkflow_run）、その成功時だけ`playback-e2e.yml`（既存のまま）という順序へ組み替えた。無効化されていた`generator-system.yml`workflowをGitHub API経由で再有効化した。

PR #176〜#178がこの間にdevelopへmergeされたため、GitHub側でPR #179のbaseがdevelopへ自動的に切り替わった。リモートの追従（daily/lessons）をmergeしてpushした。

## 2. Changes

1. `apps/diagrams/runtime.py`の原稿・TTS経路を、旧設計（Cursor primary→Gemini fallback 1段、TTS fallbackなし）から現行実装（Gemini free→Cursor→Gemini paid finalの3段、TTS Gemini free→Gemini paid finalの2段）へ更新し、PNGを再生成した
2. `DESIGN.md`§3外部I/O表の原稿・TTS行を、fallback順序を示す記述へ更新した
3. `DEPLOY.md`のcredential一覧・`generator-draft-rate.yml`/`generator-system.yml`の使う登録・rate計測workflow記述にあった`TEST_CURSOR_API_KEY`/`TEST_SPARE_GEMINI_API_KEY`等の旧登録名を、Decision 2026-09-16T00-39-21準拠の現行登録名へ揃えた
4. `docs/tasks/todo/generator-lane.md`の「済み」1〜16番（storage移行・fallback初期実装等の詳細な進捗ログ）を7項目の要約へ圧縮し、D（未決）セクションから今回の4PRで解決済みになった項目を削除した
5. T2/T3実装が参照していた不足decision（`2026-09-16T11-41-26` / `2026-09-16T13-06-32`）を元worktreeから補完した
6. `generator-system.yml`の起動条件を`pull_request: branches:[master]`から`push: branches:[master]`へ変更し、`playback-deploy.yml`を`generator-system`成功時の`workflow_run`発火へ変更した（`apps/playback/**` path絞り込みは廃止）。CD連鎖のdecisionを追加した
7. `gh api -X PUT repos/{owner}/{repo}/actions/workflows/345774169/enable`で無効化されていた`generator-system.yml`を再有効化した
8. PR #176〜#178のdevelop merge後、リモートに追加されたdaily/lessons差分をmergeしてpushした

### Commits

- `1ecef49`
- `515671f`
- `a87520d`
