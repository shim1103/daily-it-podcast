## 1. Summary

このIssueでは、gate 外の Playwright E2E で Access 入場済み session（`storageState`）を用い、同一 origin の一覧・再生・原稿の最終結果を self-validate する。完了後、週次 `playback-e2e` が placeholder ではなくアプリ回帰になる。

## 2. Context

1. E2E 入口（script / workflow）は A 済み。placeholder は実 spec に置換済み（Note §11）。
2. OTP 毎回自動化はしない。storageState を使う（Decision `2026-08-30T16-20-03`）。
3. `PLAYWRIGHT_*` 登録・取得手順の正は `DEPLOY.md`。
4. Drive / `/episodes*` は double しない（本番 Worker 経路）。
5. 正常系は本番 `DRIVE_FOLDER_ID` 直下の安定 fixture 1 件以上に寄せる（Decision `2026-08-31T00-12-01`）。空一覧は E2E の主契約にしない。安定 fixture は ProduceEpisode / TextWriter の validation（`ManuscriptDraftFromWriterOutput`）と `WriteEpisode` を通る完成成果物とする（Decision `2026-08-31T00-22-00`）。

## 3. Canonical Sources

1. Decision `docs/decisions/2026-08-30T16-20-00-docs-playback-integration-e2e-plan.md` — gate 外・定時
2. Decision `docs/decisions/2026-08-30T16-20-03-docs-playback-integration-e2e-plan.md` — OTP / storageState / Drive
3. Decision `docs/decisions/2026-08-31T00-12-01-feature-playback-e2e-deploy.md` — E2E 正常系・安定 fixture・専用 folder 非作成
4. Decision `docs/decisions/2026-08-31T00-22-00-feature-playback-e2e-deploy.md` — fixture は ProduceEpisode/TextWriter + WriteEpisode を通る
5. `DEPLOY.md` — §7 Verification・Playback E2E 登録
6. `apps/playback/playwright.config.ts` / `scripts/playback/test-e2e.sh` / `.github/workflows/playback-e2e.yml`
7. test 方針 — `testing-strategy` / playwright skill（短絡 bypass 禁止）

## 4. Scope

### In Scope

1. placeholder を本番 assert に置き換える（または併置して placeholder を外す）。
2. 認証済みで一覧・原稿・再生 UI が観測できる case（episode ≥1）。
3. `DEPLOY.md` に従い GHA Secret を想定した実行（local でも同 env で再現可能にする）。
4. 安定 fixture 原稿を `apps/playback/test/e2e/fixtures/stable-episode/` に置き、人手で本番 `DRIVE_FOLDER_ID` へ upload する前提を満たす。

### Out of Scope

1. OTP / 拒否 email の自動化（Phase 2 手動のまま）。
2. Service Token / preview URL。
3. Google OAuth / Drive UI。
4. NI / BI の再 assert。
5. 必須 Integration / Unit gate への搭載。
6. Playback E2E 専用の第3 Drive folder / 別 Worker env。

## 5. Contract

1. `PLAYWRIGHT_BASE_URL` + 有効 `storageState` があるとき、認証後に episode が **1 件以上**ある一覧が表示される（Decision `2026-08-31T00-12-01`）。
2. そのうち安定 fixture episode を開くと原稿が表示される。
3. 再生 UI（`<audio>`）が付く。
4. `./scripts/playback/test-e2e.sh` が上記前提で exit 0。
5. Secret 未設定時の振る舞いを壊して必須 gate を赤くしない（週次 job は Secret 登録後に意味を持つ）。
6. 空一覧（0 件）は E2E の主契約にしない（下位 Scope が所有）。

## 6. Constraints

1. test 専用の API bypass / in-memory 短絡を production path に足さない。
2. secret 値を log / repo に出さない。
3. generator を触らない。
4. 本番 Worker が読む `DRIVE_FOLDER_ID` を変えず、安定 fixture（`{episodeId}.json` + `{episodeId}.wav`）を人手配置する。fixture 原稿の repo 正本 path は `apps/playback/test/e2e/fixtures/stable-episode/`。

## 7. Acceptance Criteria

1. [ ] AC-1: 認証済みで episode ≥1 の一覧が見える（安定 fixture を含む）。
2. [ ] AC-2: 認証済みで安定 fixture の原稿が見える。
3. [ ] AC-3: 認証済みで再生 UI がある。
4. [ ] AC-4: `./scripts/playback/test-e2e.sh` が前提 env 付きで pass する。
5. [ ] AC-5: Integration / Unit 必須 gate に E2E を載せないままである。
6. [ ] AC-6: 空一覧を E2E 主契約にしていない（Decision `2026-08-31T00-12-01`）。

## 8. Verification

```bash
# Secret / storageState を DEPLOY.md どおり用意したうえで
# 安定 fixture を本番 DRIVE_FOLDER_ID へ人手 upload 済みであること
./scripts/playback/test-e2e.sh
./scripts/test-gate-composer-sociable-unit.shell
```

## 9. Dependencies

1. blocked by: Phase 2（本番 hostname + Access + §7 手動入場が一度できること）。storageState 取得の前提。
2. blocked by: 安定 fixture を本番 `DRIVE_FOLDER_ID` 直下へ人手配置（repo 正本: `apps/playback/test/e2e/fixtures/stable-episode/`）。
3. related: Worker BI / Drive NI（下位。E2E は再 assert しない）。

## 10. Risks

1. storageState 失効（session 30 日）— Secret 再登録手順は `DEPLOY.md`。
2. 安定 fixture 未配置 / 破損 — 一覧・原稿・再生の正常系が観測できない。空一覧を主契約に戻さない（Decision `2026-08-31T00-12-01`）。
3. 日次 produce で episode が増えても安定 fixture は残す前提。本番一覧に fixture も見えることを受け入れる。

## 11. Notes

1. GitHub Issue 化しない運用。本 file が達成契約の正。
2. follow-up: OTP 拒否の自動証明は非 scope のまま。
3. Note: placeholder は削除済み。`apps/playback/test/e2e/authenticated_playback.e2e.spec.ts` に一覧・原稿・`<audio>` の実 spec がある。`PLAYWRIGHT_BASE_URL` / `PLAYWRIGHT_STORAGE_STATE` 未設定時は skip で exit 0（必須 gate を赤くしない）。AC-1..4 の完了は deploy・安定 fixture upload・storageState 付き実行が残る。
