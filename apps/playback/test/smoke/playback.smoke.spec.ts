import { expect, test } from "@playwright/test";

/**
 * scope: Playback smoke（Access 以外の本当の e2e。deploy 前・gate 外）
 * real: Vite dev server 上で動く Hono app 一式（route・controller・use-case・composition root・
 *   GoogleDriveEpisodeRepository）。TEST 専用 Google OAuth / Drive credential で実 Google API・
 *   実 Drive へ疎通する。
 * double: なし。Access のみ持たない（origin が localhost のため Cloudflare Access を経由しない）。
 * @require TEST_GOOGLE_OAUTH_* / TEST_DRIVE_FOLDER_ID が web/vite.smoke.config.ts 経由で
 *   process.env にある（無ければ Worker 側が Runtime Config Error を返す）。
 * @ensure 一覧 API が成功応答を返し、一覧 root が描画される。TEST Drive folder は
 *   generator-system.yml の System test が都度書込・削除する運用のため、件数を固定 assert しない
 *   （0 件でも成功応答なら良い）。episode が 1 件以上あるときだけ、選択・再生・seek の操作が
 *   落ちないことを確認する。
 * 判断: docs/decisions/2026-09-15T05-36-32-chore-cd-release-protect-smoke-e2e.md
 */
test.describe("playback smoke (real Drive via TEST credential)", () => {
  test("list responds successfully without a runtime config error", async ({ page }) => {
    // Given: TEST credential 経由の実 Drive 一覧
    // When: 一覧 root を開く
    await page.goto("/");

    // Then: runtime config error（設定不足）が出ていない。一覧 root か空状態のどちらかが描画される
    const list = page.locator(".episode-list");
    const pageError = page.locator("[data-page-error]");
    await expect(list.or(pageError)).toBeVisible({ timeout: 30_000 });
    await expect(pageError).toHaveCount(0);
  });

  test("selecting the first episode (if any) opens its manuscript without error", async ({
    page,
  }) => {
    // Given: 一覧が描画された状態
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list).toBeVisible({ timeout: 30_000 });

    // When: 1 件以上あれば先頭 episode を選択する
    const firstSelect = list.locator("article.episode-item button").first();
    const hasEpisode = (await firstSelect.count()) > 0;
    test.skip(!hasEpisode, "TEST Drive folder に episode が無いため skip");
    await firstSelect.click();

    // Then: 原稿 opening が例外無く描画される
    await expect(page.locator("[data-manuscript-opening]")).toBeVisible({ timeout: 30_000 });
  });

  test("playing the first episode (if any) attaches an audio src without error", async ({
    page,
  }) => {
    // Given: 一覧が描画された状態
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list).toBeVisible({ timeout: 30_000 });

    // When: 1 件以上あれば先頭 episode の再生ボタンを押す
    const firstRow = list.locator("article.episode-item").first();
    const hasEpisode = (await firstRow.count()) > 0;
    test.skip(!hasEpisode, "TEST Drive folder に episode が無いため skip");
    await firstRow.getByRole("button", { name: "再生" }).click();

    // Then: <audio> に src が付く（実 mp3 download が例外にならない）
    const audio = page.locator(".audio-controls audio");
    await expect(audio).toBeVisible({ timeout: 30_000 });
    const src = await audio.getAttribute("src");
    expect(src).toBeTruthy();
    expect(src).toContain("/episodes/");
  });
});
