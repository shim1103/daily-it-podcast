## 1. Summary

`apps/generator/test/getxapi_narrow_integration_test.go` を追加し、processenv dummy Bearer + httptest TLS で GetXAPI Adapter の外向き境界を Integration gate（secret なし Narrow）で self-validate する。既存 sociable unit から境界 I/O 観測を移し、Unit と Narrow の責務を分離する。

## 2. Context

1. Integration gate は `scripts/generator/test-integration.sh` → `go test ./test/...`
2. Adapter は `Inject{Bearer}` で Authorization 付き HTTP を生成する
3. `post_source_sociable_unit_test.go` が processenv + DialTLSContext + httptest を含み、Unit/Narrow が混在している

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
2. `testing-strategy/levels.md` / `credential.md`
3. `apps/generator/internal/infrastructure/x/getxapi/post_source.go`
4. `apps/generator/internal/infrastructure/secrettransport/processenv/client.go`
5. `apps/generator/internal/infrastructure/x/getxapi/post_source_sociable_unit_test.go`（移管元）

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/getxapi_narrow_integration_test.go`（build tag なし）
2. Narrow へ移す観測: processenv dummy、DialTLSContext redirect、Authorization（Bearer prefix）、成功 1 item、unresolved binding で未到達
3. Unit に残す観測: nil client、HTTP 非 200 等の Adapter 内分岐（境界 I/O なし）
4. Unit から移管 case を削除し二重検証をやめる

### Out of Scope

1. AgentSecrets / `local_real` / vendor 実 API / Broad・System・E2E

## 5. Contract

1. Integration gate exit 0
2. Authorization 非空かつ Bearer prefix、成功時 SourceItem 1 件以上
3. error message に dummy token を含めない
4. Unit = Adapter 内分岐のみ。Narrow = 境界 I/O のみ

## 6. Constraints

1. dummy 値を failure message に出さない
2. `local_real` / AgentSecrets を使わない

## 7. Acceptance Criteria

- [ ] Narrow file が CI gate で収集・実行できる
- [ ] upstream 到達と Authorization 存在
- [ ] 成功 1 item 以上と失敗分岐の区別
- [ ] unresolved binding で upstream 未到達
- [ ] **責務分離**: sociable unit から processenv+httptest 観測を Narrow へ移し、Unit に残さない
- [ ] **責務分離**: Unit は Adapter 内分岐のみ。Narrow は境界 I/O のみ

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/x/getxapi/ -count=1
```

## 9. Dependencies

- `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`

## 10. Risks

1. DialTLSContext host mismatch → 「呼ばれた/呼ばれてない」のみ観測

## 11. Notes

GitHub Issue 化しない。本 file が契約の正。
