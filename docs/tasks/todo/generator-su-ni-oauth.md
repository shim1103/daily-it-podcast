## 1. Summary

このIssueでは、Google OAuth TokenSourceのSociable UnitとNarrow Integrationをtarget architecture（`*http.Client` + capability Config）へ揃え、Adapter内分岐とtoken endpoint境界I/Oの責務を分離する。完了後はUnitが境界I/Oを持たず、secretなしNarrowがrefresh grantの外向き契約をself-validateする。

## 2. Context

1. 旧`generator-narrow-gate-vendor-oauth.md`はprocessenv / `secrettransport`前提だったため、本fileへ統合して置き換える。
2. 現行`token_source_sociable_unit_test.go`はprocessenv + DialTLS + httptest観測と分岐検証が混在している。
3. M1完了後の配線を正とし、processenv形Narrowを先行実装しない。

## 3. Canonical Sources

1. `docs/decisions/2026-08-28T12-49-00-docs-infra-unit-narrow-integration-latest.md` — M1後にSU/NIを揃える順序。
2. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md` — Integration gateはsecretなしNarrow。
3. `docs/decisions/2026-08-27T13-56-14-docs-env-secret-management-reconsider.md` — HTTP Adapter依存形。
4. `apps/generator/internal/infrastructure/google/oauth/` — TokenSourceとUnit test。
5. `testing-strategy` — levels / credential / 二重最小化。

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/oauth_narrow_integration_test.go`（build tagなし）。
2. Narrow観測: token endpointへのPOST到達、refresh grantに必要なform、成功時access token。
3. Sociable Unitから境界I/O成功経路を除き、nil client / 401 / 空token / 不正JSON等の分岐を残す。
4. 同一observableの二重検証をやめる。

### Out of Scope

1. Composition再結線と`secrettransport`削除（M1）。
2. Google実API、Broad / System・E2E。
3. gdrive writer本体のSU/NI（`generator-su-ni-gdrive.md`）。

## 5. Contract

1. Integration gateが本Narrowを含みexit 0。
2. formと成功時access tokenが期待どおり。
3. failure messageにdummy credential値を出さない。
4. Unit = Adapter内分岐のみ。Narrow = 境界I/Oのみ。

## 6. Constraints

1. dummy値をfailure messageに出さない。
2. `secrettransport` / processenv Clientを再導入しない。
3. GitHub Issue化しない。本fileが契約の正。

## 7. Acceptance Criteria

1. [ ] Narrow fileがCI gateで収集・実行できる。
2. [ ] 成功時POST formとaccess tokenが期待どおり。
3. [ ] 失敗時OAuth Infrastructure Errorを識別できる（Unit側で足りるならNarrowは最小）。
4. [ ] Sociable Unitにprocessenv + httptest成功経路観測が残っていない。
5. [ ] UnitはAdapter内分岐のみ、Narrowは境界I/Oのみ。
6. [ ] error messageにdummy値がない。

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/google/oauth/ -count=1
```

## 9. Dependencies

1. `docs/tasks/todo/generator-composition-http-adapters.md`（M1）完了後。

## 10. Risks

1. form名ずれ時は「request到達」を優先観測し、secret値を出さない。

## 11. Notes

1. 旧narrow-gate taskの観測意図（form・access token・責務分離）は本fileが引き継ぐ。
