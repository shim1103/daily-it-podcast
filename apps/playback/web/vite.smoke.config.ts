import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { createApp } from "../worker/src/routes/app.ts";
import { createDevBackendMiddleware } from "./dev-backend-proxy.ts";
import { createRemoteTestR2Binding } from "../test/support/create-remote-test-r2-binding.ts";

/**
 * Playback smoke 専用 dev server。origin は localhost:3000（web/vite.config.ts の dummy backend と同じ形）。
 *
 * `createApp()` を `{ mode: "r2" }` 明示で呼び、`env.EPISODES` に実 TEST R2 binding
 * （`test/support/create-remote-test-r2-binding.ts`、bucket は `daily-it-podcast-dev`）を注入する。
 * 本番 route（`routes/app.ts`）が R2 mode 固定になったため、smoke も同じ経路（R2 正本）を確認する。
 *
 * @require 実 Cloudflare 認証（`CLOUDFLARE_API_TOKEN` 等）が process env にある（無ければ
 *   getPlatformProxy の remoteBindings 起動が失敗する）
 * @invariant 本番 bucket（`daily-it-podcast-prod`）には触らない。web/vite.config.ts（fixture 固定の
 *   開発用 dummy backend）は変更しない
 * 判断: docs/decisions/2026-09-15T05-36-32-chore-cd-release-protect-smoke-e2e.md
 */
export default defineConfig({
  root: new URL(".", import.meta.url).pathname,
  server: {
    port: 3000,
    middlewareMode: false,
  },
  plugins: [
    react(),
    {
      name: "smoke-backend-api",
      async configureServer(server) {
        const { bucket, dispose } = await createRemoteTestR2Binding();
        const smokeApp = createApp();

        server.middlewares.use(
          createDevBackendMiddleware(async (req: Request): Promise<Response> => {
            return smokeApp.fetch(req, { EPISODES: bucket });
          }),
        );

        server.httpServer?.once("close", () => {
          void dispose();
        });
      },
    },
  ],
});
