import { defineConfig, devices } from "@playwright/test";

/**
 * Playback browser E2E 入口。
 * remote（Access 付き本番 hostname）前提。webServer は必須にしない。
 * env の意味・GHA 写像の正は DEPLOY.md「Playback E2E 登録」。
 * 判断: docs/decisions/2026-08-30T16-20-00 / 2026-08-30T16-20-03
 */
export default defineConfig({
  testDir: "test/e2e",
  fullyParallel: true,
  reporter: "list",
  use: {
    ...devices["Desktop Chrome"],
    // optional: PLAYWRIGHT_BASE_URL（本番 workers.dev origin）。未設定時は remote e2e を skip。
    baseURL: process.env.PLAYWRIGHT_BASE_URL,
    // optional: PLAYWRIGHT_STORAGE_STATE（storageState JSON の path）。未設定時は remote e2e を skip。
    storageState: process.env.PLAYWRIGHT_STORAGE_STATE || undefined,
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
