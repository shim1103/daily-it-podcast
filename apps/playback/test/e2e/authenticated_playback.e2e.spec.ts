import { expect, test } from "@playwright/test";

/** 安定 fixture（UI は date のみ。episodeId は DOM に出ない）。値は fixtures/stable-episode/*.json を写す */
const FIXTURE_DATE_UI = "2026/09/04";
const FIXTURE_EPISODE_ID = "8ff4177b-26fe-4036-ab7b-d2a4e9e7639d";
const FIXTURE_OPENING_SNIPPET = "2026年9月4日";
/** fixture topics[0].startSec。seek 先が 0:00 に落ちない回帰確認に使う */
const FIXTURE_FIRST_TOPIC_START_SEC = 38.04;

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

  test("seeking to the first topic moves audio currentTime to that topic's startSec, not 0", async ({
    page,
  }) => {
    // Given: 安定 fixture を選択し原稿を開く（再生は押さない）
    await page.goto("/");
    const list = page.locator(".episode-list");
    await expect(list.locator("[data-episode-title]").first()).toBeVisible({ timeout: 60_000 });
    const fixtureRow = list
      .locator("article.episode-item")
      .filter({ has: page.locator("[data-episode-date]", { hasText: FIXTURE_DATE_UI }) });
    await fixtureRow.locator("button").first().click();

    // When: 先頭 topic の seek bar を押す（音源未 load 状態からの seek）
    const firstTopicSeek = page.locator(".episode-topic [data-topic-start-sec]").first();
    await expect(firstTopicSeek).toBeVisible({ timeout: 60_000 });
    await firstTopicSeek.click();

    // Then: <audio> の currentTime が topic の startSec 付近へ動く（0:00 に落ちない）
    const audio = page.locator(".audio-controls audio");
    await expect(audio).toBeVisible({ timeout: 60_000 });
    await expect
      .poll(async () => audio.evaluate((el: HTMLAudioElement) => el.currentTime), {
        timeout: 30_000,
      })
      .toBeGreaterThan(FIXTURE_FIRST_TOPIC_START_SEC - 5);
  });
});
