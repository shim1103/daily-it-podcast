## 1. Summary

`apps/generator/test/cursorcli_narrow_integration_test.go` を追加し、processenv dummy Cursor key + 実 child（fake Cursor CLI）で TextWriter の command 境界を Integration gate（secret なし Narrow）で self-validate する。既存 sociable unit から「実プロセス境界」に属する観測を移し、Unit と Narrow の責務を分離する。

## 2. Context

1. TextWriter は `commandlaunch.Launcher` 経由で Cursor CLI を起動し、stdout JSON の `result` を fragment にする
2. `text_writer_sociable_unit_test.go` は Launcher Stub で envelope decode を検証している（ここは Unit のまま妥当）
3. 本 task は processenv launcher + 実 child 起動の境界を Narrow として gate に載せる。Unit に processenv 実起動を足さない

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
2. `testing-strategy/levels.md` / `credential.md`
3. `apps/generator/internal/infrastructure/manuscript/cursorcli/text_writer.go`
4. `apps/generator/internal/infrastructure/manuscript/cursorcli/constants.go`
5. `apps/generator/internal/infrastructure/commandlaunch/processenv/launcher.go`
6. `apps/generator/internal/infrastructure/manuscript/cursorcli/text_writer_sociable_unit_test.go`（Unit 側の正）
7. `apps/generator/test/commandlaunch_processenv_narrow_integration_test.go`（processenv 境界の既存 Narrow）

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/cursorcli_narrow_integration_test.go`（build tag なし）
2. Narrow 観測: fake `BinaryName` を PATH に置き、processenv launcher + TextWriter で実 child 起動、成功 fragment、失敗時 Infrastructure Error・断片空、secret 値が error/stdout に出ない
3. Unit に残す観測: Launcher Stub による envelope decode、argv/stdin 形、nil/空 brief（実 child / processenv なし）
4. Unit に processenv 実起動や fake binary PATH 観測を追加しない。二重検証しない

### Out of Scope

1. real Cursor CLI / AgentSecrets / `local_real` / System・E2E

## 5. Contract

1. Integration gate exit 0
2. 成功時 fragment 返却。失敗時 Infrastructure Error・断片空
3. secret 値を failure message に出さない
4. Unit = Stub 上の Adapter 内分岐。Narrow = processenv + 実 child 境界

## 6. Constraints

1. dummy Cursor API key 値を failure message に出さない
2. process stdout 全量をログ化しない

## 7. Acceptance Criteria

- [ ] Narrow file が CI gate で収集・実行できる
- [ ] 成功時 TextWriter.Write が fragment を返す
- [ ] 失敗時 Infrastructure Error・断片空
- [ ] failure message に secret 値なし
- [ ] **責務分離**: Unit は Launcher Stub のまま。processenv 実起動・fake binary 観測は Narrow のみ
- [ ] **責務分離**: Narrow は envelope 分岐の網羅を持たない（Unit に残す）。境界 I/O の最小成功/失敗のみ

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/manuscript/cursorcli/ -count=1
```

## 9. Dependencies

- `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
- processenv command launcher

## 10. Risks

1. envelope 形式 mismatch → Unit の envelope() 形に合わせ、`result` のみ必須

## 11. Notes

GitHub Issue 化しない。本 file が契約の正。
