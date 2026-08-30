## 1. Summary

このIssueでは、secret なし Narrow Integration で Google Drive HTTP 境界の実 I/O 契約を self-validate する。完了後、Integration gate が当該 suite を収集し、SU の Stub `fetch` では検出できない境界失敗を落とせる。

## 2. Context

1. Integration gate は secret なし NI + BI（Decision `2026-08-30T16-20-00`）。
2. 既存 `google-drive-episode-repository.sociable_unit.test.ts` は Stub `fetch` であり、levels 上は SU。
3. Unit coverage 分母は SU + secret なし NI（Decision `2026-08-30T16-20-01`）。
4. E2E / Access は本 Issue の対象外。

## 3. Canonical Sources

1. Decision `docs/decisions/2026-08-30T16-20-00-docs-playback-integration-e2e-plan.md` — gate
2. Decision `docs/decisions/2026-08-30T16-20-01-docs-playback-integration-e2e-plan.md` — coverage 分母
3. `apps/playback/worker/src/infrastructure/drive/google-drive-episode-repository.ts` — production Adapter
4. generator Narrow（`apps/generator/test/gdrive_narrow_integration_test.go` 等）— 実 HTTP 形の参照
5. `DESIGN.md` §5 — 置き場・命名
6. test 方針 — `testing-strategy`（levels / minimization / naming-and-layout）

## 4. Scope

### In Scope

1. `apps/playback/test/` に `*_narrow_integration*.test.ts` を追加し、Drive HTTP 1境界を実物にする。
2. Sociable Unit から「実 I/O 契約」を自称する観測を外し、Adapter 内分岐だけを SU に残す。
3. secret 実値を使わない（dummy + 実 TLS/HTTP 形）。

### Out of Scope

1. Broad / E2E。
2. API Client Narrow。
3. SU が所有する分岐網羅の再 assert。
4. 本番 / TEST Drive folder への credential 付き実 operation。

## 5. Contract

1. Integration gate が当該 Narrow file を収集・実行し exit 0。
2. 実 HTTP 経路で list / download 相当の成功と代表失敗（非 2xx・形式不正等）が DriveError 等の既存契約へ写る。
3. error / log に dummy secret が漏れない。

## 6. Constraints

1. Stub / Spy だけの suite を Narrow と呼ばない（Decision `2026-08-30T16-20-02`）。
2. 上位 Scope が所有する内部詳細を再 assert しない。
3. generator を触らない。

## 7. Acceptance Criteria

1. [ ] AC-1: Narrow file が Integration gate で収集・実行される。
2. [ ] AC-2: Unit gate（coverage）が secret なし NI を分母に含めたまま pass する。
3. [ ] AC-3: SU は Stub `fetch` の実 I/O 自称をやめ、Adapter 内分岐に閉じる。
4. [ ] AC-4: `./scripts/playback/test-integration.sh` と `./scripts/playback/test-unit.sh` が pass する。

## 8. Verification

```bash
./scripts/playback/check-static.sh
./scripts/playback/test-unit.sh
./scripts/playback/test-integration.sh
```

test 方針は `testing-strategy` を参照する。

## 9. Dependencies

1. related: 既存 Drive SU（責務分離の対象）。
2. related: Worker BI 正常系 Issue（本 Issue の境界 I/O を奪わない）。

## 10. Risks

1. SU と NI の二重観測 — minimization に従い観測を移す。
2. happy-dom 環境で実 HTTP が足りない — Narrow は node 環境へ寄せる判断を実装時に取る。

## 11. Notes

1. GitHub Issue 化しない運用（playback-lane）。本 file が達成契約の正。
