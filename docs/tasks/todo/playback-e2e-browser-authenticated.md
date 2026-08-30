## 1. Summary

このIssueでは、gate 外の Playwright E2E で Access 入場済み session（`storageState`）を用い、同一 origin の一覧・再生・原稿の最終結果を self-validate する。完了後、週次 `playback-e2e` が placeholder ではなくアプリ回帰になる。

## 2. Context

1. E2E 入口（script / workflow / placeholder）は A 済み。
2. OTP 毎回自動化はしない。storageState を使う（Decision `2026-08-30T16-20-03`）。
3. `PLAYWRIGHT_*` 登録・取得手順の正は `DEPLOY.md`。
4. Drive / `/episodes*` は double しない（本番 Worker 経路）。

## 3. Canonical Sources

1. Decision `docs/decisions/2026-08-30T16-20-00-docs-playback-integration-e2e-plan.md` — gate 外・定時
2. Decision `docs/decisions/2026-08-30T16-20-03-docs-playback-integration-e2e-plan.md` — OTP / storageState / Drive
3. `DEPLOY.md` — §7 Verification・Playback E2E 登録
4. `apps/playback/playwright.config.ts` / `scripts/playback/test-e2e.sh` / `.github/workflows/playback-e2e.yml`
5. test 方針 — `testing-strategy` / playwright skill（短絡 bypass 禁止）

## 4. Scope

### In Scope

1. placeholder を本番 assert に置き換える（または併置して placeholder を外す）。
2. 認証済みで一覧・原稿・再生 UI が観測できる case。
3. `DEPLOY.md` に従い GHA Secret を想定した実行（local でも同 env で再現可能にする）。

### Out of Scope

1. OTP / 拒否 email の自動化（Phase 2 手動のまま）。
2. Service Token / preview URL。
3. Google OAuth / Drive UI。
4. NI / BI の再 assert。
5. 必須 Integration / Unit gate への搭載。

## 5. Contract

1. `PLAYWRIGHT_BASE_URL` + 有効 `storageState` があるとき、一覧が表示される。
2. episode を開くと原稿が表示される。
3. 再生 UI（`<audio>`）が付く。
4. `./scripts/playback/test-e2e.sh` が上記前提で exit 0。
5. Secret 未設定時の振る舞いを壊して必須 gate を赤くしない（週次 job は Secret 登録後に意味を持つ）。

## 6. Constraints

1. test 専用の API bypass / in-memory 短絡を production path に足さない。
2. secret 値を log / repo に出さない。
3. generator を触らない。

## 7. Acceptance Criteria

1. [ ] AC-1: 認証済みで一覧が見える。
2. [ ] AC-2: 認証済みで原稿が見える。
3. [ ] AC-3: 認証済みで再生 UI がある。
4. [ ] AC-4: `./scripts/playback/test-e2e.sh` が前提 env 付きで pass する。
5. [ ] AC-5: Integration / Unit 必須 gate に E2E を載せないままである。

## 8. Verification

```bash
# Secret / storageState を DEPLOY.md どおり用意したうえで
./scripts/playback/test-e2e.sh
./scripts/test-gate-composer-sociable-unit.shell
```

## 9. Dependencies

1. blocked by: Phase 2（本番 hostname + Access + §7 手動入場が一度できること）。storageState 取得の前提。
2. related: Worker BI / Drive NI（下位。E2E は再 assert しない）。

## 10. Risks

1. storageState 失効（session 30 日）— Secret 再登録手順は `DEPLOY.md`。
2. Drive 空 folder — 一覧 0 件でも「表示できる」契約にするか fixture episode を要るかは実装時に観測して決める（本 Issue で AC を壊さない範囲）。

## 11. Notes

1. GitHub Issue 化しない運用。本 file が達成契約の正。
2. follow-up: OTP 拒否の自動証明は非 scope のまま。
