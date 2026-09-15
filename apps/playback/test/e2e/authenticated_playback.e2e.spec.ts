import { expect, test } from "@playwright/test";

/**
 * 安定 fixture（UI は date のみ。episodeId は DOM に出ない）。値は fixtures/stable-episode/*.json を写す
 */
const FIXTURE_DATE_UI = "2026/09/04";

const hasRemoteAuth =
  Boolean(process.env.PLAYWRIGHT_BASE_URL?.trim()) &&
  Boolean(process.env.PLAYWRIGHT_STORAGE_STATE?.trim());

/**
 * scope: Playback browser E2E（Access session 付き・deploy 後・gate 外）
 * @invariant 一覧・原稿・再生という UI 機能は smoke（test/smoke/playback.smoke.spec.ts、実 browser・
 *   TEST 専用 credential 経由の実 Drive・deploy 前）が既に検証する。ここで同じ機能を再assertしない
 *   （minimization.md §2-4・levels.md §6-1-5: smokeとE2Eはdeploy前後で検証可能な範囲が変わる境界で分担する）。
 *   E2Eが所有するのは「Access sessionを経由して本番Driveへ実際に到達できるか」という、
 *   smokeでは検証不可能な最終結果1点のみ。
 * 判断: docs/decisions/2026-09-15T05-36-32-chore-cd-release-protect-smoke-e2e.md
 */
test.describe("authenticated playback remote e2e", () => {
  test.skip(!hasRemoteAuth, "PLAYWRIGHT_BASE_URL / PLAYWRIGHT_STORAGE_STATE 未設定のため skip");

  test("reaches the production Drive through Access session and shows the stable fixture date", async ({
    page,
  }) => {
    // Given: Access 入場済み storageState と本番 baseURL
    // When: 一覧 root を開く
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list).toBeVisible({ timeout: 60_000 });
    const titles = list.locator("[data-episode-title]");
    await expect(titles.first()).toBeVisible({ timeout: 60_000 });

    // Then: episode ≥1 かつ安定 fixture 日付が並ぶ（Access を経由して本番 Drive へ実際に到達できた証）
    expect(await titles.count()).toBeGreaterThanOrEqual(1);
    await expect(list.locator("[data-episode-date]", { hasText: FIXTURE_DATE_UI })).toBeVisible();
  });
});
