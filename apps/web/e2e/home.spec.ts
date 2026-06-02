import { test, expect } from '@playwright/test';

test.describe('Home page — エピソード一覧', () => {
  test('Given ホームページにアクセス When ページが読み込まれる Then h1 と EpisodeList が表示される', async ({ page }) => {
    await page.goto('/');

    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('h1')).toContainText('Daily IT Podcast');

    // MockDriveService の固定エピソードが表示される
    await expect(page.locator('a[href*="/episode/"]')).toBeVisible();
  });

  test('Given エピソードリンクが存在する When リンクをクリック Then episode ページに遷移する', async ({ page }) => {
    await page.goto('/');

    const firstLink = page.locator('a[href*="/episode/"]').first();
    await expect(firstLink).toBeVisible();
    await firstLink.click();

    await expect(page).toHaveURL(/\/episode\//);
  });
});
