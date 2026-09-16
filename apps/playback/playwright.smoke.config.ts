import { defineConfig, devices } from "@playwright/test";

/**
 * Playback smoke 入口（Access 以外の本当の e2e。deploy 前・gate 外）。
 * `vite --config web/vite.smoke.config.ts` を webServer として起動し、実 browser 経由で一覧・原稿・
 * 再生・seek を、TEST 専用 Google OAuth / Drive credential 経由の実 Drive データで確認する。
 * Access だけを持たない構成であり、それ以外は本番相当の合成経路を実際に通す。
 * 判断: docs/decisions/2026-09-15T05-36-32-chore-cd-release-protect-smoke-e2e.md
 */
export default defineConfig({
  testDir: "test/smoke",
  fullyParallel: false,
  reporter: "list",
  webServer: {
    command: "npm run dev:smoke",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
  use: {
    ...devices["Desktop Chrome"],
    baseURL: "http://localhost:3000",
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
