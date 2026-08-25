## 1. Summary

この Issue では、Generator の HTTP Adapter 用 `processenv` transport を実装し、Adapter を `secrettransport.Client` 依存へ切り替えて Composition 結線する。完了後、production は process environment から必要な secret だけを HTTP へ注入できる。

## 2. Context

1. HTTP credential の契約と production 置き場（process environment）は Decision で固定済み。
2. HTTP Adapter は現状 `agentsecrets.Client` 具象に直接依存している。
3. Cursor command / child process は別出口であり本 Issue の対象外。
4. local AgentSecrets HTTP 実装は後続 Issue（`generator-agentsecrets-http-transport.md`）。本 Issue は Adapter 切替と processenv 実装まで。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/secrettransport/contract.go` — HTTP transport 契約。
2. `apps/generator/internal/composition/secret_bindings.go` — `SecretRef` binding。
3. `docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` — Adapter と runtime の責務境界。
4. `docs/decisions/2026-08-22T18-44-00-feature-generator-infras-all-narrow-integration.md` — production process environment。
5. `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md` — 置き場×出口。
6. `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md` — Issue 分割と依存。
7. test 方針 — `testing-strategy`。

## 4. Scope

### In Scope

1. `secrettransport.Client` を満たす process-env HTTP implementation。
2. HTTP Adapter を `secrettransport.Client` 依存へ切り替え、Composition で processenv を結線すること。
3. process environment を実境界とする HTTP Narrow Integration。

### Out of Scope

1. Cursor command launcher / child process environment。
2. local AgentSecrets HTTP 実装の完成と実 proxy 検証（`generator-agentsecrets-http-transport.md`）。
3. `cmd/generator`、production GHA workflow、実 vendor API。

## 5. Contract

1. `SecretRef` を Composition binding で解決し、header、form、JSON field へ注入する。
2. 未設定または解決不能な secret reference は外部 HTTP 前に失敗する。
3. secret 値を error または log へ出さない。
4. HTTP Adapter は `secrettransport.Client` に依存し、`agentsecrets.Client` 具象に依存しない。

## 6. Constraints

1. Generator code は GitHub Actions 固有 API を import しない。
2. AgentSecrets HTTP runtime は削除せず、後続 Issue で選ぶ別 implementation として残せる形を壊さない。
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

1. blocked by: `secrettransport` 契約と production processenv 決定（完了）。
2. blocks: `generator-agentsecrets-http-transport.md`。
3. related: process-env Cursor command launcher（完了）。local Cursor command（独立）。

## 10. Risks

1. HTTP request または error への secret 値混入 → dummy value の Narrow Integration で注入と非露出を検証する。

## 11. Notes

1. local AgentSecrets HTTP は本 Issue では「消さない」まで。実装・結線は follow-up Issue。
2. 決定本文は Canonical Sources の Decision を参照し、ここへ転記しない。
