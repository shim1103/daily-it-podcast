---
name: R2疎通確認dispatch実行とmasterへの最小到達、実装削除、Drive→R2人手移行の一時展開
date: 2026-09-16T13:50:00
session_id: none
branch: feature/r2-smoke-remove-and-migrate
prev: なし
---

## 1. Summary

列6（`r2-smoke-migrate-cutover`）の疎通確認dispatchをmasterへ到達させ実行しPASSさせた後、一過性の疎通確認実装（yml・smoke.go等）を削除した。Drive→R2の人手移行対象データ（test/prod）を`/tmp`へ展開し、Issue docsをDRY原則に合わせて整理した。

## 2. Changes

1. `generator-r2-smoke.yml`をmasterへ最小PR（file単体、`apps/playback`を含めずdeploy誤発火を回避）で到達させた
2. `--ref`明示dispatchでR2疎通確認（Put→List→Get往復）をtest bucketで実行しPASSを確認した（run 35054835880）
3. 一過性の疎通確認実装（yml・smoke.go・関連test・script）を削除した
4. Drive→R2人手移行対象のzip（test_episodes・episodes）を`/tmp/tests/`・`/tmp/prod/`へ展開した
5. Issue file（`r2-smoke-migrate-cutover.md`）の削除理由が3箇所（Scope・AC・Notes）に分散していたDRY違反を、新規Decisionへの一元化で解消した
6. shimがprod/test両bucketへ手動でR2移行を実施（put完了）。stem数・pair整合の確認はまだ行っていない前提で、列6の残タスク（人手移行確認・本番同着deploy・DEPLOY.md更新）を列7（`r2-post-cutover-verify-oauth`）へ統合し、列6のIssue fileを完了削除した

### Commits

- `4b84935`（masterへのyml単体PR、別branch `feature/generator-r2-smoke-yml-only`）
- `a603a17`
- `5fd84ee`
- `4e08b7b`
