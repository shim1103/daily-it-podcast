## 1. Summary

このIssueでは、secret なし Broad Integration で Worker 入口から Composition 経由の正常系（一覧・詳細・音声）配線を self-validate する。既存の設定不足 BI と役割分担し、Narrow では検出できない合成失敗を Integration gate で落とす。

## 2. Context

1. 既存 BI は `runtime_config_boundary.broad_integration.test.ts`（設定不足 → configuration_error）のみ。
2. BI 入口は `workerEntry.fetch` 仮定（browser からは送らない）。
3. Drive は double（真外部）。production Composition は通す。
4. E2E / Access は対象外。

## 3. Canonical Sources

1. Decision `docs/decisions/2026-08-30T16-20-00-docs-playback-integration-e2e-plan.md` — gate
2. 既存 `apps/playback/test/runtime_config_boundary.broad_integration.test.ts`
3. `apps/playback/worker/src/worker-entry.ts` / Composition Root
4. `apps/playback/contracts/`
5. `DESIGN.md` §5
6. test 方針 — `testing-strategy`（levels / minimization）
7. generator Broad 達成契約 — helper 共有の形の参照（`docs/tasks/todo/generator-broad-integration-produce-episode.md`）

## 4. Scope

### In Scope

1. `apps/playback/test/` に正常系 `*_broad_integration*.test.ts` を追加する。
2. list / get episode / get audio が Worker 入口から届くこと（代表 case）を assert する。
3. 代表的な error 伝播が合成で初めて見える範囲だけを追加する。

### Out of Scope

1. Narrow の境界 I/O 再 assert。
2. SU の分岐網羅。
3. E2E・Access・frontend Page BI。
4. credential 付き実 Drive。

## 5. Contract

1. 設定十分な env + Drive double で、list / get / audio の成功応答が入口から返る。
2. 既存 config 不足 BI の契約を壊さない。
3. Integration gate で収集・実行され exit 0。

## 6. Constraints

1. minimization: 下位 Scope の内部詳細を再 assert しない。
2. browser / Playwright を起動しない。
3. generator を触らない。

## 7. Acceptance Criteria

1. [ ] AC-1: 正常系 Broad file が Integration gate で収集・実行される。
2. [ ] AC-2: list 成功が Worker 入口から観測できる。
3. [ ] AC-3: get episode 成功が Worker 入口から観測できる。
4. [ ] AC-4: get audio 成功が Worker 入口から観測できる。
5. [ ] AC-5: 既存 config 不足 BI が引き続き pass する。
6. [ ] AC-6: `./scripts/playback/test-integration.sh` が pass する。

## 8. Verification

```bash
./scripts/playback/check-static.sh
./scripts/playback/test-unit.sh
./scripts/playback/test-integration.sh
```

## 9. Dependencies

1. related: Drive Narrow（境界 I/O の所有者。本 Issue は奪わない）。
2. blocked by: なし（Drive double で進められる）。

## 10. Risks

1. Narrow helper への逆依存 — 共有が必要なら中立 support へ出す。

## 11. Notes

1. GitHub Issue 化しない運用。本 file が達成契約の正。
