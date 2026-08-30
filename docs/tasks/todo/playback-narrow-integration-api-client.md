## 1. Summary

このIssueでは、secret なし Narrow Integration で playback API Client の真 HTTP 境界を self-validate する。完了後、Stub `fetch` の SU では検出できない wire / 実到達の失敗を Integration gate で落とせる。

## 2. Context

1. 既存 `playback-api-client.sociable_unit.test.ts` は `fetch` Stub（Decision `2026-08-30T16-20-02` どおり SU）。
2. Narrow は実 network（local listener 等）へ届くことだけを追加で見る。
3. browser E2E・Access は対象外。

## 3. Canonical Sources

1. Decision `docs/decisions/2026-08-30T16-20-02-docs-playback-integration-e2e-plan.md` — SU / NI 境界
2. Decision `docs/decisions/2026-08-30T16-20-00-docs-playback-integration-e2e-plan.md` — gate
3. `apps/playback/web/src/api/playback-api-client.ts` / `playback-rpc-client.ts`
4. `apps/playback/contracts/` — HTTP schema
5. `DESIGN.md` §5
6. test 方針 — `testing-strategy`

## 4. Scope

### In Scope

1. `apps/playback/test/` に API Client 向け `*_narrow_integration*.test.ts` を追加する。
2. 実 HTTP で list / get の成功と代表 status → Result 写像を検証する。
3. schema 細部の網羅は SU / contracts SU に残し、Narrow は到達と代表写像に限る。

### Out of Scope

1. Stub `fetch` の SU 網羅の再 assert。
2. Worker Composition BI / Drive Narrow / E2E。
3. Playwright。

## 5. Contract

1. 実 TCP/HTTP で契約 path に届き、成功時は schema 準拠 Result、代表失敗は既存 error 語彙へ写る。
2. Integration gate で収集・実行され exit 0。

## 6. Constraints

1. Stub / Spy のみを Narrow と呼ばない。
2. minimization: SU 所有の分岐を再 assert しない。
3. generator を触らない。

## 7. Acceptance Criteria

1. [ ] AC-1: Narrow file が Integration gate で収集・実行される。
2. [ ] AC-2: 実 HTTP 成功経路で list または get が ok Result になる。
3. [ ] AC-3: 代表非成功 status が既存 API error 語彙へ写る。
4. [ ] AC-4: `./scripts/playback/test-integration.sh` が pass する。

## 8. Verification

```bash
./scripts/playback/check-static.sh
./scripts/playback/test-unit.sh
./scripts/playback/test-integration.sh
```

## 9. Dependencies

1. related: Drive Narrow（別境界。混ぜない）。
2. related: Worker BI（Worker 入口側。本 Issue は web API Client 側）。

## 10. Risks

1. local listener の寿命・port — test 内で閉じ、外部 port 固定に依存しない。

## 11. Notes

1. GitHub Issue 化しない運用。本 file が達成契約の正。
