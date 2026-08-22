## 1. Summary

このIssueでは、Generator の HTTP Adapter 用 `processenv` transport を実装する。production runtime が Composition の `SecretRef` binding を使い、必要な secret だけを HTTP request へ注入できる状態にする。

## 2. Context

1. HTTP credential runtime の contract は A で固定済み。
2. production credential runtime は process environment を使うことを決定済み。
3. HTTP request 注入と Cursor child process 起動は別外部境界であり、後者は対象外。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/secrettransport/contract.go` — HTTP transport contract。
2. `apps/generator/internal/composition/secret_bindings.go` — secret reference binding。
3. `docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` — Adapter と runtime の責務境界。
4. `docs/decisions/2026-08-22T18-44-00-feature-generator-infras-all-narrow-integration.md` — production process environment runtime の決定。
5. `testing-strategy` — test Scope と secret の扱い。

## 4. Scope

### In Scope

1. `secrettransport.Client` を満たす process-env HTTP implementation。
2. HTTP Adapter を A contract へ接続する Composition 結線。
3. process environment を実境界とする HTTP Narrow Integration。

### Out of Scope

1. Cursor command launcher と child process environment。
2. local AgentSecrets の実 proxy 検証。
3. `cmd/generator`、production GHA workflow、実 vendor API。

## 5. Contract

1. `SecretRef` を Composition binding で解決し、header、form、JSON field へ注入する。
2. 未設定または解決不能な secret reference は外部 HTTP 前に失敗する。
3. secret 値を error または log へ出さない。

## 6. Constraints

1. Generator code は GitHub Actions 固有 API を import しない。
2. AgentSecrets HTTP runtime は置換せず、Composition が選択する別 implementation として追加する。
3. test は dummy process environment を使う。

## 7. Acceptance Criteria

1. [ ] HTTP Adapter が concrete AgentSecrets client でなく `secrettransport.Client` に依存する。
2. [ ] process-env HTTP transport が header、form、JSON field を注入する。
3. [ ] 未設定 secret は network 前に失敗し、secret 値を error / log に出さない。
4. [ ] HTTP Narrow Integration が dummy process environment と test upstream で contract を self-validate する。
5. [ ] Generator の static、Unit、Integration gate が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-integration.sh
```

## 9. Dependencies

A の HTTP transport contract と B の production runtime decision は完了済み。

## 10. Risks

HTTP request または error への secret 値の混入。dummy value を使う Narrow Integration で注入と非露出を検証する。

## 11. Notes

Cursor command launch は `generator-processenv-command-launcher.md` が所有する。
