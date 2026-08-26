## 1. Summary

この Issue では、Cursor CLI 用 local AgentSecrets command launcher を `commandlaunch.Launcher` として実装し、Composition から選択結線できる状態にする。完了後、local では Cursor 専用 project 経由で起動でき、production の `processenv` 結線は残る。

## 2. Context

1. production Cursor path は `processenv.Launcher` 結線済み。Composition から project を外したのは processenv path の YAGNI であり、local 専用 project 思想の撤回ではない。
2. `agentsecrets.EnvWrapper` と `CursorProjectName` は `commandlaunch/agentsecrets` に残るが Composition 未結線。
3. Issue 分割と依存は `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md` を正とする。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/commandlaunch/contract.go` — command launch 契約。
2. `apps/generator/internal/infrastructure/commandlaunch/agentsecrets/env_wrapper.go` — CLI wrapper / project dir 規約。
3. `apps/generator/internal/composition/secret_bindings.go` — Composition binding / allowlist 所有。
4. `docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md` — CLI は `agentsecrets env --`。
5. `docs/decisions/2026-08-22T15-08-00-agentsecrets-cursor-project-template-boundary.md` — Cursor 専用 project。
6. `docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` — Composition が project / runtime 選択を所有。
7. `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md` — 置き場×出口、結線と契約の区別。
8. `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md` — local Issue は出口軸で分離。
9. test 方針 — `testing-strategy`。

## 4. Scope

### In Scope

1. `commandlaunch.Launcher` を満たす AgentSecrets（Cursor project）実装。
2. Composition が Cursor 専用 project・allowlist を所有し、local 選択時にその launcher を結線できること。
3. production `processenv` 結線の維持。

### Out of Scope

1. process-env / AgentSecrets の HTTP transport。
2. production 既定を AgentSecrets に戻すこと。
3. `cmd/generator`、GHA workflow、実 Cursor API、専用 project の運用作成自動化。

## 5. Contract

1. Cursor Adapter は `commandlaunch.Launcher` のみに依存する。
2. local launcher は Composition が渡す Cursor 専用 project で注入範囲を閉じる。
3. child 継承 env は Composition allowlist に従い、親環境を全継承しない。
4. `Command` に secret・project・runtime kind を載せない。error / log に secret 値・stdin・child stderr 本文を出さない。

## 6. Constraints

1. Generator code は GitHub Actions 固有 API を import しない。
2. `processenv.Launcher` を削除・置換しない。
3. Cursor 専用 project を repo 内へ移さない。
4. test は dummy project dir / test child を使い、本番 keychain を読まない。

## 7. Acceptance Criteria

1. [ ] AgentSecrets 系 `commandlaunch.Launcher` 実装が存在する。
2. [ ] Composition が Cursor 専用 project を所有し、local 選択時にその launcher を結線できる。
3. [ ] production `processenv` path の既存 gate / Narrow Integration が退行しない。
4. [ ] Cursor Adapter が AgentSecrets 具象へ戻らない。
5. [ ] Generator の static、Unit、Integration gate が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-integration.sh
```

## 9. Dependencies

1. blocked by: process-env Cursor command launcher（完了）。
2. related: `generator-agentsecrets-http-transport.md`（出口が違うため互いに block しない）。
3. related: `generator-processenv-http-transport.md`（HTTP production。本 Issue と独立）。

## 10. Risks

1. local を production 既定に誤って戻す → Composition の選択境界と verification で production path 退行を見る。
2. project dir 不備で注入範囲が広がる → 絶対 path / 専用 project 契約を test で固定する。

## 11. Notes

1. follow-up の HTTP local は `generator-agentsecrets-http-transport.md`。
2. 決定本文は Canonical Sources の Decision を参照し、ここへ転記しない。
