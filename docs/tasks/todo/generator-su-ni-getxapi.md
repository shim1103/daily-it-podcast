## 1. Summary

このIssueでは、GetXAPI AdapterのSociable UnitとNarrow Integrationをtarget architecture（`*http.Client` + capability Config）へ揃え、Adapter内分岐と外向きHTTP境界I/Oの責務を分離する。完了後はUnitが境界I/Oを持たず、secretなしNarrowがIntegration gateでBearer到達をself-validateする。

## 2. Context

1. 旧`generator-narrow-gate-vendor-getxapi.md`はprocessenv / `secrettransport`前提だったため、本fileへ統合して置き換える。
2. 現行`post_source_sociable_unit_test.go`は境界I/O観測とAdapter分岐が混在している。
3. M1完了後の配線を正とし、processenv形Narrowを先行実装しない。

## 3. Canonical Sources

1. `docs/decisions/2026-08-28T12-49-00-docs-infra-unit-narrow-integration-latest.md` — M1後にSU/NIを揃える順序。
2. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md` — Integration gateはsecretなしNarrow。
3. `docs/decisions/2026-08-27T13-56-14-docs-env-secret-management-reconsider.md` — HTTP Adapter依存形。
4. `apps/generator/internal/infrastructure/x/getxapi/` — AdapterとUnit test。
5. `testing-strategy` — levels / credential / 二重最小化。

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/getxapi_narrow_integration_test.go`（build tagなし）。
2. Narrow観測: Controllable upstream（httptest等）への到達、Authorization（Bearer）、成功時SourceItem最小、失敗時未到達またはInfrastructure Error。
3. Sociable Unitから境界I/O観測を除き、Adapter内分岐だけを残す。
4. 同一observableの二重検証をやめる。

### Out of Scope

1. Composition再結線と`secrettransport`削除（M1）。
2. vendor実API、Broad / System・E2E。
3. 他vendor（oauth / gemini / gdrive / cursor）のSU/NI。

## 5. Contract

1. Integration gateが本Narrowを含みexit 0。
2. 成功時Authorizationが非空かつBearer、SourceItemが1件以上。
3. failure messageにdummy credential値を出さない。
4. Unit = Adapter内分岐のみ。Narrow = 境界I/Oのみ。

## 6. Constraints

1. dummy値をfailure message / probe logに出さない。
2. `secrettransport` / processenv Clientを再導入しない。
3. GitHub Issue化しない。本fileが契約の正。

## 7. Acceptance Criteria

1. [ ] Narrow fileがCI gateの`go test ./test/...`で収集・実行できる。
2. [ ] upstream到達とAuthorization（Bearer）を観測できる。
3. [ ] 成功時SourceItemが1件以上返る。
4. [ ] Sociable Unitにhttptest / DialTLS実到達観測が残っていない。
5. [ ] UnitはAdapter内分岐のみ、Narrowは境界I/Oのみ。
6. [ ] error messageにdummy token値がない。

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/x/getxapi/ -count=1
```

## 9. Dependencies

1. `docs/tasks/todo/generator-composition-http-adapters.md`（M1）完了後。

## 10. Risks

1. host redirect不一致で「0 calls」だけ見える場合、secret値は出さず到達有無だけ報告する。

## 11. Notes

1. 旧narrow-gate taskの観測意図（Bearer・成功item・責務分離）は本fileが引き継ぐ。
