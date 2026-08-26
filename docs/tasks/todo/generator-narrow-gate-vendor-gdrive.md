## 1. Summary

`apps/generator/test/gdrive_narrow_integration_test.go` を追加し、processenv dummy Folder ID + httptest TLS で GDrive Writer の外向き境界を Integration gate（secret なし Narrow）で self-validate する。既存 sociable unit から境界 I/O 観測を移し、Unit と Narrow の責務を分離する。

## 2. Context

1. Writer は `secrettransport.Client` + TokenSource で Drive API へ list/create/upload する
2. `writer_sociable_unit_test.go` が processenv + DialTLSContext + httptest で request shape を検証しており、Unit/Narrow が混在している

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
2. `testing-strategy/levels.md` / `credential.md`
3. `apps/generator/internal/infrastructure/drive/gdrive/writer.go`
4. `apps/generator/internal/infrastructure/secrettransport/processenv/client.go`
5. `apps/generator/internal/infrastructure/drive/gdrive/writer_sociable_unit_test.go`（移管元）

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/gdrive_narrow_integration_test.go`（build tag なし）
2. Narrow へ移す観測: processenv Folder ID 注入、TokenSource stub、DialTLSContext redirect、成功ルート 1 連（list/create/upload）、json/wav stem 整合、Authorization
3. Unit に残す観測: 入力 validation、Infrastructure Error 種別、既存 file update 分岐など Adapter 内論理（境界 I/O なし、または Client Stub のみ）
4. Unit から移管 case を削除し二重検証をやめる

### Out of Scope

1. Google 実 API / AgentSecrets / Broad・System・E2E

## 5. Contract

1. Integration gate exit 0
2. probe が成功 call sequence を 1 連受ける。json/wav stem 整合
3. error message に dummy 値を含めない
4. Unit = Adapter 内分岐のみ。Narrow = 境界 I/O のみ

## 6. Constraints

1. dummy 値を failure message に出さない
2. `local_real` / AgentSecrets を使わない

## 7. Acceptance Criteria

- [ ] Narrow file が CI gate で収集・実行できる
- [ ] upstream が成功 call sequence を 1 連受ける
- [ ] json/wav stem 整合
- [ ] error message に dummy 値なし
- [ ] **責務分離**: sociable unit から processenv+httptest 観測を Narrow へ移し、Unit に残さない
- [ ] **責務分離**: Unit は Adapter 内分岐のみ。Narrow は境界 I/O のみ

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/drive/gdrive/ -count=1
```

## 9. Dependencies

- `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`

## 10. Risks

1. path 判定ずれ → method 到達の観測に最小化

## 11. Notes

GitHub Issue 化しない。本 file が契約の正。
