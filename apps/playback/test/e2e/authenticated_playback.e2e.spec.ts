import { expect, test } from "@playwright/test";

/** 安定 fixture（UI は date のみ。episodeId は DOM に出ない） */
const FIXTURE_DATE_UI = "2026/08/30";
const FIXTURE_EPISODE_ID = "eb14426e-2f4c-4157-9175-301ae4e7808d";
const FIXTURE_OPENING_SNIPPET = "2026年8月30日";

const hasRemoteAuth =
  Boolean(process.env.PLAYWRIGHT_BASE_URL?.trim()) &&
  Boolean(process.env.PLAYWRIGHT_STORAGE_STATE?.trim());

test.describe("authenticated playback remote e2e", () => {
  test.skip(!hasRemoteAuth, "PLAYWRIGHT_BASE_URL / PLAYWRIGHT_STORAGE_STATE 未設定のため skip");

  test("shows at least one episode including the stable fixture date on the list", async ({
    page,
  }) => {
    // Given: Access 入場済み storageState と本番 baseURL
    // When: 一覧 root を開く
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list).toBeVisible({ timeout: 60_000 });
    const titles = list.locator("[data-episode-title]");
    await expect(titles.first()).toBeVisible({ timeout: 60_000 });

    // Then: episode ≥1 かつ安定 fixture 日付が並ぶ
    expect(await titles.count()).toBeGreaterThanOrEqual(1);
    await expect(list.locator("[data-episode-date]", { hasText: FIXTURE_DATE_UI })).toBeVisible();
  });

  test("opens the stable fixture and shows manuscript opening and topics", async ({ page }) => {
    // Given: 認証済み一覧
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list.locator("[data-episode-title]").first()).toBeVisible({ timeout: 60_000 });

    // When: 安定 fixture 行の select ボタン（Row 先頭 button）を押す
    const fixtureRow = list
      .locator("article.episode-item")
      .filter({ has: page.locator("[data-episode-date]", { hasText: FIXTURE_DATE_UI }) });
    await fixtureRow.locator("button").first().click();

    // Then: opening に日付文言があり、topic が見える
    const opening = page.locator("[data-manuscript-opening]");
    await expect(opening).toBeVisible({ timeout: 60_000 });
    await expect(opening).toContainText(FIXTURE_OPENING_SNIPPET);
    await expect(page.locator("[data-topic-title]").first()).toBeVisible();
  });

  test("shows audio controls with src pointing at the fixture episode", async ({ page }) => {
    // Given: 認証済み一覧から安定 fixture の再生ボタンを押す
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list.locator("[data-episode-title]").first()).toBeVisible({ timeout: 60_000 });
    const fixtureRow = list
      .locator("article.episode-item")
      .filter({ has: page.locator("[data-episode-date]", { hasText: FIXTURE_DATE_UI }) });
    // why: 新 UI では audio は再生（Row の 2 個目の button）で現れる。select では現れない
    await fixtureRow.getByRole("button", { name: "再生" }).click();

    // Then: <audio controls> があり src が /episodes/{id} を含む
    const audio = page.locator(".audio-controls audio");
    await expect(audio).toBeVisible({ timeout: 60_000 });
    await expect(audio).toHaveAttribute("controls", "");
    const src = await audio.getAttribute("src");
    expect(src).toBeTruthy();
    expect(src).toContain("/episodes/");
    expect(src).toContain(FIXTURE_EPISODE_ID);
  });
});
