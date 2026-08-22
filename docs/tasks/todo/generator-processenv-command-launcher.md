## 1. Summary

このIssueでは、Generator の Cursor CLI 用 `processenv` command launcher を実装する。production runtime が Composition の Cursor binding に従い、必要な child environment だけを渡して command を起動できる状態にする。

## 2. Context

1. command launch contract と Cursor runtime binding は A で固定済み。
2. production credential runtime は process environment を使うことを決定済み。
3. HTTP request 注入と child process 起動は別外部境界であり、前者は対象外。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/commandlaunch/contract.go` — command launch contract。
2. `apps/generator/internal/composition/secret_bindings.go` — Cursor project、secret reference、child env allowlist の正。
3. `docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` — Adapter と runtime の責務境界。
4. `docs/decisions/2026-08-22T18-44-00-feature-generator-infras-all-narrow-integration.md` — production process environment runtime の決定。
5. `testing-strategy` — test Scope と secret の扱い。

## 4. Scope

### In Scope

1. `commandlaunch.Launcher` を満たす process-env implementation。
2. Cursor Adapter を A contract へ接続する Composition 結線。
3. 実 child process を境界にする command launcher Narrow Integration。

### Out of Scope

1. HTTP secret transport。
2. local AgentSecrets wrapper の実境界検証。
3. `cmd/generator`、production GHA workflow、実 Cursor API 呼び出し。

## 5. Contract

1. Cursor Adapter が作る `Command` に secret、project dir、runtime kind を含めない。
2. process-env launcher は Composition binding の allowlist と Cursor secret だけを child process へ渡す。
3. 未設定 secret または空の program は child process 起動前に失敗する。
4. secret 値、stdin、child stderr 本文を error または log へ出さない。

## 6. Constraints

1. Generator code は GitHub Actions 固有 API を import しない。
2. AgentSecrets exec wrapper は置換せず、Composition が選択する別 implementation として追加する。
3. test は dummy process environment と test child process を使う。

## 7. Acceptance Criteria

1. [ ] Cursor Adapter が concrete AgentSecrets wrapper でなく `commandlaunch.Launcher` に依存する。
2. [ ] process-env launcher が allowlist と Cursor secret だけを child process へ渡す。
3. [ ] 未設定 secret または空 program は child process 起動前に失敗し、secret 値・stdin・stderr 本文を error / log に出さない。
4. [ ] command launcher Narrow Integration が dummy process environment と test child process で contract を self-validate する。
5. [ ] Generator の static、Unit、Integration gate が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-integration.sh
```

## 9. Dependencies

A の command launch contract と B の production runtime decision は完了済み。

## 10. Risks

child environment への過剰継承と error への secret / stderr 混入。dummy value と test child process で allowlist と非露出を検証する。

## 11. Notes

HTTP secret transport は `generator-processenv-http-transport.md` が所有する。
