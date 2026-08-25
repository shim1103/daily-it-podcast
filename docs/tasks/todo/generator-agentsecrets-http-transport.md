## 1. Summary

この Issue では、HTTP Adapter 用 local AgentSecrets runtime を `secrettransport.Client` 実装として追加し、Composition から選択結線できる状態にする。完了後、local では proxy 経由の秘密注入を契約経由で使え、production の processenv HTTP 結線は残る。

## 2. Context

1. HTTP Adapter は現状 `agentsecrets.Client` 具象に直接依存している。最終系では `secrettransport.Client` のみに依存する。
2. production HTTP（processenv 実装 + Adapter 切替）は別 Issue が所有する。本 Issue はその後に、local 実装を差し込む。
3. Issue 分割と依存は `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md` を正とする。
4. 仮定: production HTTP Issue 完了時点で Adapter は `secrettransport.Client` 依存になっている。

## 3. Canonical Sources

1. `apps/generator/internal/infrastructure/secrettransport/contract.go` — HTTP transport 契約。
2. `apps/generator/internal/infrastructure/agentsecrets/proxy.go` — AgentSecrets HTTP proxy Client。
3. `apps/generator/internal/composition/secret_bindings.go` — `SecretRef` binding。
4. `docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` — Adapter と runtime の責務境界。
5. `docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md` — HTTP 境界は proxy。
6. `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md` — 置き場×出口。
7. `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md` — local Issue は出口軸で分離。
8. test 方針 — `testing-strategy`。

## 4. Scope

### In Scope

1. `secrettransport.Client` を満たす AgentSecrets HTTP（proxy）実装。
2. Composition が local 選択時にその Client を HTTP Adapter へ結線できること。
3. production processenv HTTP 結線の維持。

### Out of Scope

1. process-env HTTP transport 本体と HTTP Adapter の初回 `secrettransport` 切替（`generator-processenv-http-transport.md`）。
2. Cursor command / AgentSecrets exec wrapper / Cursor 専用 project。
3. `cmd/generator`、GHA workflow、実 vendor API、実 proxy 常駐の運用自動化。

## 5. Contract

1. HTTP Adapter は `secrettransport.Client` のみに依存する（本 Issue 開始時点でそうなっている前提。崩れていれば production HTTP Issue へ戻す）。
2. local Client は Composition binding の `SecretRef` を解決し、header / form / JSON field へ注入する。
3. 未設定または解決不能な secret は外部 HTTP 前に失敗する。
4. secret 値を error / log に出さない。

## 6. Constraints

1. Generator code は GitHub Actions 固有 API を import しない。
2. processenv HTTP 実装を削除・置換しない。
3. test は dummy / test double を使い、本番 keychain と実 upstream を必須にしない。実 proxy 検証は本 Issue の必須 AC にしない。

## 7. Acceptance Criteria

1. [ ] AgentSecrets 系 `secrettransport.Client` 実装が存在する。
2. [ ] Composition が local 選択時にその Client を HTTP Adapter へ結線できる。
3. [ ] production processenv HTTP path が退行しない。
4. [ ] HTTP Adapter が `agentsecrets.Client` 具象へ戻らない。
5. [ ] Generator の static、Unit、Integration gate が pass する。

## 8. Verification

```bash
./scripts/generator/check-static.sh
./scripts/generator/test-unit.sh
./scripts/generator/test-integration.sh
```

## 9. Dependencies

1. blocked by: `generator-processenv-http-transport.md`（Adapter が `secrettransport.Client` 依存になること）。
2. related: `generator-agentsecrets-cursor-command-launcher.md`（互いに block しない）。

## 10. Risks

1. production HTTP 未完了のまま着手し具象依存が残る → Dependencies の blocked by を守る。
2. secret 値の error / log 混入 → contract と Narrow / Unit で非露出を見る。

## 11. Notes

1. Cursor 専用 project は HTTP には使わない。command 側 Issue の所有。
2. 決定本文は Canonical Sources の Decision を参照し、ここへ転記しない。
