import { test, expect } from '@playwright/test';

const MOCK_EPISODE_ID = 'mock-episode-001';

test.describe('ManuscriptViewer — MeaningPopup', () => {
  test('Given 原稿テキストを選択 When mouseup イベント発生 Then MeaningPopup が表示される', async ({ page }) => {
    // /api/meaning を intercept してモックレスポンスを返す
    await page.route('/api/meaning', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ meaning: 'TypeScript は型安全な JavaScript の superset です。' }),
      });
    });

    await page.goto(`/episode/${MOCK_EPISODE_ID}`);

    // 原稿セクションのテキストを選択する
    const manuscriptSection = page.locator('section').first();
    await expect(manuscriptSection).toBeVisible();

    // テキストをダブルクリックして単語を選択し mouseup をトリガー
    const openingText = page.getByText('本日のITニュースをお届けします。');
    await openingText.dblclick();

    // MeaningPopup: role="dialog" が表示される
    await expect(page.locator('[role="dialog"]')).toBeVisible();
  });

  test('Given MeaningPopup が表示されている When 閉じるボタンをクリック Then Popup が非表示になる', async ({ page }) => {
    await page.route('/api/meaning', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ meaning: 'モックの意味テキストです。' }),
      });
    });

    await page.goto(`/episode/${MOCK_EPISODE_ID}`);

    // テキストを選択して Popup を開く
    const openingText = page.getByText('本日のITニュースをお届けします。');
    await openingText.dblclick();

    await expect(page.locator('[role="dialog"]')).toBeVisible();

    // 閉じるボタンをクリック
    await page.locator('[aria-label="閉じる"]').click();

    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
  });
});
