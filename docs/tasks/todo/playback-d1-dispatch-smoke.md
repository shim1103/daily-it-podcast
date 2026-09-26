# feat(playback): 進捗D1の一過性疎通（workflow_dispatch）とPASS後削除

## 1. Summary

このIssueでは、Aで登録済みの TEST D1 に対し一過性の `workflow_dispatch` 疎通を通し、PASS後に疎通専用artifactを削除する。常設smoke／E2E拡張は別Issue。

## 2. Context

1. 事実: test/prod D1 instance・`database_id`・`wrangler.smoke.jsonc` 結線・GHA登録はA済み。
2. 事実: R2では疎通dispatchと常設gateを分け、PASS後に疎通専用実装を削除した（Decision `2026-09-16T13-46-41`）。
3. 仮定: 表が無いと疎通が落ちる場合、dispatch内でmigration適用してよい（登録のやり直しではない）。

## 3. Canonical Sources

1. A: `apps/playback/wrangler.jsonc` / `test/support/wrangler.smoke.jsonc` の `EPISODE_PROGRESS`・`EPISODE_PROGRESS_D1_BINDING`
2. B: `docs/decisions/2026-09-16T13-46-41-feature-r2-smoke-remove-and-migrate.md`（一過性PASS後削除）
3. B: `docs/decisions/2026-09-16T10-45-14-feature-r2-smoke-migrate-cutover.md`（疎通は生I/O往復のみ）
4. scope-split §5（一過性疎通C）
5. test方針: `skills` の `testing-strategy`

## 4. Scope

### In Scope

1. TEST D1向け一過性 `workflow_dispatch`（＋最小probe）
2. prepare／簡単な write+read 往復のPASS
3. PASS後の疎通専用artifact削除

### Out of Scope

1. Port／adapter本実装、list embed、web
2. 常設 `playback-smoke`／System／E2EへのD1載せ（release Issue）
3. Fake／local peer本実装
4. instance／credentialの再登録（A済み）

## 5. Contract

変更しない。疎通は登録済みbinding名・TEST instanceを使う。

## 6. Constraints

1. test固定。prod切替引数を持たない。
2. merge意味・HTTP応答形・UIは見ない。
3. 常設gate拡張を同一Issueに混ぜない。

## 7. Acceptance Criteria

1. [ ] A登録済みTEST D1で疎通がPASSする
2. [ ] 疎通専用artifact（dispatch workflow・一時script等）をPASS後に削除する
3. [ ] 常設smoke／e2e ymlをこのIssueで拡張していない

## 8. Verification

1. `workflow_dispatch` をTEST向けに実行しPASSを確認する
2. 削除後、repoに疎通専用入口が残っていないことを確認する

## 9. Dependencies

独立可（A登録済み前提）。

## 10. Risks

1. Secret不足で落ちる → Verificationで要否を確定し、不足分だけ登録（値の百科はDEPLOY）

## 11. Notes

follow-up: 常設smokeへD1は `playback-progress-release-verify`
