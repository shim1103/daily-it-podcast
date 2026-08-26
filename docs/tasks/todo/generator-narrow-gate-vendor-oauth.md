## 1. Summary

`apps/generator/test/oauth_narrow_integration_test.go` を追加し、processenv dummy（client id/secret/refresh）+ httptest TLS で OAuth TokenSource の外向き境界を Integration gate（secret なし Narrow）で self-validate する。既存 sociable unit から境界 I/O 観測を移し、Unit と Narrow の責務を分離する。

## 2. Context

1. TokenSource は `secrettransport.Client` で token endpoint へ refresh grant を送る
2. `token_source_sociable_unit_test.go` が processenv + DialTLSContext + httptest で form probe しており、Unit/Narrow が混在している

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
2. `testing-strategy/levels.md` / `credential.md`
3. `apps/generator/internal/infrastructure/google/oauth/token_source.go`
4. `apps/generator/internal/infrastructure/secrettransport/processenv/client.go`
5. `apps/generator/internal/infrastructure/google/oauth/token_source_sociable_unit_test.go`（移管元）

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/oauth_narrow_integration_test.go`（build tag なし）
2. Narrow へ移す観測: processenv dummy 注入、DialTLSContext redirect、POST form（`grant_type=refresh_token` 等）、成功時 access token
3. Unit に残す観測: nil client、401/空 token/不正 JSON、unresolved binding の Error 種別（境界 I/O を持たない形へ整理）
4. Unit から移管 case を削除し二重検証をやめる

### Out of Scope

1. Google 実 API / AgentSecrets / Broad・System・E2E

## 5. Contract

1. Integration gate exit 0
2. form fields 期待どおり。成功時 access token 返却
3. error message に dummy 値なし
4. Unit = Adapter 内分岐のみ。Narrow = 境界 I/O のみ

## 6. Constraints

1. dummy 値を failure message に出さない
2. `local_real` / AgentSecrets を使わない

## 7. Acceptance Criteria

- [ ] Narrow file が CI gate で収集・実行できる
- [ ] 成功時 POST form と access token が期待どおり
- [ ] 失敗時 OAuth Infrastructure Error を識別できる（Unit 側で足りるなら Narrow は最小）
- [ ] error message に dummy 値なし
- [ ] **責務分離**: sociable unit から processenv+httptest 成功経路観測を Narrow へ移し、Unit に残さない
- [ ] **責務分離**: Unit は Adapter 内分岐のみ。Narrow は境界 I/O のみ

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/google/oauth/ -count=1
```

## 9. Dependencies

- `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`

## 10. Risks

1. form 名ずれ → 「request 到達」を優先観測

## 11. Notes

GitHub Issue 化しない。本 file が契約の正。
