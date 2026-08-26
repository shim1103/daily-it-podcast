## 1. Summary

`apps/generator/test/twitterapiio_narrow_integration_test.go` を追加し、processenv dummy + httptest TLS で TwitterAPI.io Adapter の外向き境界を Integration gate（secret なし Narrow）で self-validate する。既存 sociable unit から境界 I/O 観測を移し、Unit と Narrow の責務を分離する。

## 2. Context

1. Integration gate は `scripts/generator/test-integration.sh` → `go test ./test/...`。`local_real` は tag 除外。
2. Adapter は `secrettransport.Client` + `Inject{Headers}` で外向き HTTP を生成する。
3. 現状 `post_source_sociable_unit_test.go` が processenv + DialTLSContext + httptest を含んでおり、Unit と Narrow が混在している。

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
2. `testing-strategy/levels.md` / `credential.md`
3. `apps/generator/internal/infrastructure/x/twitterapiio/post_source.go`
4. `apps/generator/internal/infrastructure/secrettransport/processenv/client.go`
5. `apps/generator/internal/infrastructure/x/twitterapiio/post_source_sociable_unit_test.go`（移管元）

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/twitterapiio_narrow_integration_test.go`（build tag なし）
2. Narrow へ移す観測: processenv dummy 注入、DialTLSContext redirect、upstream method/path/query、`X-API-Key` 存在、成功時 1 item decode、unresolved binding で upstream 未到達
3. Unit に残す観測: nil client、vendor status error 等の Adapter 内分岐（httptest/processenv なし、または Client Stub のみ）
4. Unit から移管した case を sociable unit 側で削除し、同一 observable の二重検証をやめる

### Out of Scope

1. AgentSecrets / `local_real` / vendor 実 API / Broad・System・E2E

## 5. Contract

1. `bash scripts/generator/test-integration.sh` が本 Narrow を含み exit 0
2. upstream 到達と `X-API-Key` 非空、成功時 SourceItem 1 件以上
3. 失敗 error message に dummy secret 値を含めない
4. Unit は境界 I/O を持たない。Narrow は Adapter 内分岐の網羅を持たない

## 6. Constraints

1. dummy 値を failure message / probe log に出さない
2. `local_real` / AgentSecrets を使わない

## 7. Acceptance Criteria

- [ ] Narrow file が CI gate の `go test ./test/...` で収集・実行できる
- [ ] upstream probe が 1 回以上到達し、`X-API-Key` が存在する
- [ ] 成功時 SourceItem が期待どおり（最小 1 item）
- [ ] unresolved binding で upstream 未到達かつ error
- [ ] **責務分離**: sociable unit から processenv+httptest 観測 case を Narrow へ移し、Unit に同 observable を残さない
- [ ] **責務分離**: Unit は Adapter 内分岐のみ。Narrow は境界 I/O（dummy processenv + httptest）のみ

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/x/twitterapiio/ -count=1
```

## 9. Dependencies

- `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`

## 10. Risks

1. DialTLSContext redirect mismatch → probe「0 calls」のみ報告し、secret 値は出さない

## 11. Notes

CI gate 対象の secret なし Narrow。GitHub Issue 化しない。本 file が契約の正。
