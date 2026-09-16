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
7. GitHub master branch protectionを更新（`required_pull_request_reviews`追加、review 1件必須。`enforce_admins`は`false`のまま維持）
8. developへのリリースPR（#172、develop→master）作成中に、masterへの直接push（`b5599ed`、cron時刻を19:00 JSTへ変更）がdevelopに未反映と判明。shimの判断で05:00 JSTをnew policyとして確定しdevelopへ取り込んだ
9. PR #172でconflict発生。原因はmasterのPR #163 Revert（R2 CompletedEpisodeLookup取り消し）で、developとの間で17件以上のfile（`writer.go`等の本実装がstub版に巻き戻る自動merge誤り、一時的hack tool群の復活、`docs/daily/`・`docs/decisions/`の誤削除）が影響を受けていた。develop側の内容で完全一致するmerge commitを作成し解消した
10. 本番route（`app.ts`）のR2 mode固定化（PR #169）で`playback-smoke.yml`が常時`configuration_error`になっていたことを発見。fakeへ後退させず、`getPlatformProxy`の`remoteBindings`で実TEST R2 bucket（`daily-it-podcast-dev`）へ疎通する構成へ置き換え、CI上で`smoke: pass`を確認した

### Commits

- `4b84935`（masterへのyml単体PR、別branch `feature/generator-r2-smoke-yml-only`）
- `a603a17`
- `5fd84ee`
- `4e08b7b`
- `cafc87c` → `75a0892`（developへ直接反映）
- `baed516`（masterとのconflict解消merge、developへfast-forward）
- `38e80ec` → `94aa34f`（developへ直接反映）
