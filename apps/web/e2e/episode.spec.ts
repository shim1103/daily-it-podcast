import { test, expect } from '@playwright/test';

const MOCK_EPISODE_ID = 'mock-episode-001';

test.describe('Episode page — 再生画面', () => {
  test('Given episode ページにアクセス When ページが読み込まれる Then AudioPlayer と ManuscriptViewer が表示される', async ({ page }) => {
    await page.goto(`/episode/${MOCK_EPISODE_ID}`);

    // AudioPlayer: <audio> 要素が存在する
    await expect(page.locator('audio')).toBeVisible();

    // ManuscriptViewer: テキスト選択ヒントが表示される
    await expect(page.getByText('テキストを選択すると意味を検索できます')).toBeVisible();

    // 原稿の opening テキストが存在する
    await expect(page.getByText('本日のITニュースをお届けします。')).toBeVisible();
  });

  test('Given episode ページ When タイトルが表示される Then h1 にエピソードタイトルが含まれる', async ({ page }) => {
    await page.goto(`/episode/${MOCK_EPISODE_ID}`);

    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('h1')).toContainText('2026年1月1日のITニュース');
  });

  test('Given episode ページ When 「← 一覧に戻る」をクリック Then ホームページに戻る', async ({ page }) => {
    await page.goto(`/episode/${MOCK_EPISODE_ID}`);

    await page.getByText('← 一覧に戻る').click();

    await expect(page).toHaveURL('/');
  });
});
