## 1. Summary

このIssueでは、Generatorのlocal AgentSecrets経路と`local_real` test入口を現行artifactから除き、credential付き実operationをGitHub Actions runnerだけに限定した方針へ揃える。完了後は通常のlocal開発と自動testがAgentSecrets、OS keychain、local secretを要求しない。

## 2. Context

1. GeneratorでAgentSecretsを採用しないDecisionは確定している。
2. 現在はAgentSecrets用HTTP／command実装、Compositionのlocal constructor、Narrow test、`local_real` build tagと専用scriptが残っている。
3. AgentSecrets実物検証を要求する保留taskとlocal workflow文書も現行方針に反している。

## 3. Canonical Sources

1. `docs/decisions/2026-08-27T12-17-00-docs-env-secret-management-reconsider.md` — local AgentSecretsを採用せず、credential付き実operationをGitHub Actionsへ限定するDecision。
2. `DESIGN.md` — Generatorのlatest runtime config境界とtest方針。
3. `apps/generator/internal/infrastructure/commandlaunch/contract.go` — AgentSecrets削除後も維持するcommand launch契約。
4. `apps/generator/internal/infrastructure/secrettransport/contract.go` — 後続移行まで維持する現行HTTP transport契約。
5. `testing-strategy` — test scope、credential、test doubleの共通規則。

## 4. Scope

### In Scope

1. AgentSecrets固有のHTTP／command実装とtestを削除する。
2. CompositionからAgentSecrets用local constructor、project、secret key解決を削除する。
3. `local_real` build tag、専用Integration入口、関連するgate composer検査を削除する。
4. AgentSecrets実物検証を要求する保留taskとlocal workflow文書を削除する。
5. laneとlatest test説明を、local secret経路が存在しない状態へ更新する。

### Out of Scope

1. `secrettransport`全体とprocess environment実装の削除。
2. HTTP AdapterとCursor launcherのtarget architectureへの移行。
3. TwitterAPI.io artifactの削除。
4. GitHub Actionsのproduction workflowと実service検証。
5. 過去のDecision Recordとdaily recordの変更または削除。

## 5. Contract

1. Generatorのproduction command／HTTP経路は、本Issue完了時点では既存process environment実装を維持する。
2. default Unit／Integration gateは実credentialとlocal secret storeを必要としない。
3. 現行のcode、script、workflow、task入口からAgentSecretsと`local_real`の選択肢を除く。
4. 過去時点の採用理由を持つ履歴artifactは保持する。

## 6. Constraints

1. AgentSecrets削除と同時に`secrettransport`全体を削除しない。HTTP Adapter移行は別Issueの依存を持つため。
2. local実operationの代替として`.env`系fileや別secret storeを追加しない。
3. credential値をtest、log、Error、docsへ記録しない。

## 7. Acceptance Criteria

1. [ ] AgentSecrets固有のHTTP／command packageとそのtestがcurrent treeに存在しない。
2. [ ] CompositionにAgentSecrets用local constructor、project dir、secret key解決が存在しない。
3. [ ] `local_real` build tag、専用Integration script、gate composerのlocal-real契約が存在しない。
4. [ ] AgentSecrets実物検証を要求する2件の保留taskとlocal workflow文書が存在しない。
5. [ ] default Generator Unit／Integration testがlocal secretなしでpassする。
6. [ ] 過去のDecision Recordとdaily recordは変更されていない。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-race.sh
./scripts/generator/test-integration.sh
rg 'AgentSecrets|agentsecrets|local_real|test-integration-local' apps/generator scripts .github README.md DESIGN.md
test ! -e docs/tasks/todo/generator-narrow-local-agentsecrets-http.md
test ! -e docs/tasks/todo/generator-narrow-local-agentsecrets-command.md
test ! -e .agent/workflows/agentsecrets.md
git diff --check
```

`rg`はmatchなしでexit 1になることを確認する。

## 9. Dependencies

1. `docs/decisions/2026-08-27T12-17-00-docs-env-secret-management-reconsider.md`の確定済みDecisionに依存する。
2. 本Issueは`generator-remove-twitterapiio.md`より先に完了し、共有するCompositionとbinding周辺の変更を直列化する。

## 10. Risks

1. process environment production経路まで誤って削除すると後続移行前にGeneratorの結線が壊れるため、AgentSecrets固有artifactだけを削除する。
2. gate composerからlocal入口だけを除く時に他の集約契約を弱めるriskがあるため、composer testと全Generator gateを実行する。

## 11. Notes

1. `secrettransport`廃止とCursor child environment再設計は後続Issueで扱う。
