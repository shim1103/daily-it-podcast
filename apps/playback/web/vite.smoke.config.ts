import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { createApp } from "../worker/src/routes/app.ts";
import { createDevBackendMiddleware } from "./dev-backend-proxy.ts";

/**
 * Playback smoke 専用 dev server。origin は localhost:3000（web/vite.config.ts の dummy backend と同じ形）。
 *
 * `createApp()`（override 無し）を、`env` に TEST 専用 Google OAuth / Drive credential を注入して呼ぶ。
 * `createPlaybackControllers` は env が揃うと "drive" mode（本物の GoogleDriveEpisodeRepository）を
 * 選ぶため、Worker 専用 middleware も wrangler dev も要らない（Hono app の fetch(req, env) を
 * Vite dev server 上で直接呼ぶだけで済む）。
 *
 * @require TEST_GOOGLE_OAUTH_CLIENT_ID / TEST_GOOGLE_OAUTH_CLIENT_SECRET / TEST_GOOGLE_OAUTH_REFRESH_TOKEN /
 *   TEST_DRIVE_FOLDER_ID が process.env にある（無ければ Worker 側が Runtime Config Error を返し、
 *   一覧は失敗応答になる）。`TEST_` 接頭の変数名だけを読むため、無印の本番 env 変数が存在しても
 *   このプロセスには渡らない
 * @invariant 本番 credential は使わない。web/vite.config.ts（fixture 固定の開発用 dummy backend）は
 *   変更しない
 * 判断: docs/decisions/2026-09-15T05-36-32-chore-cd-release-protect-smoke-e2e.md
 */
export function createSmokeBackendMiddleware() {
  const env = {
    GOOGLE_OAUTH_CLIENT_ID: process.env.TEST_GOOGLE_OAUTH_CLIENT_ID,
    GOOGLE_OAUTH_CLIENT_SECRET: process.env.TEST_GOOGLE_OAUTH_CLIENT_SECRET,
    GOOGLE_OAUTH_REFRESH_TOKEN: process.env.TEST_GOOGLE_OAUTH_REFRESH_TOKEN,
    DRIVE_FOLDER_ID: process.env.TEST_DRIVE_FOLDER_ID,
  };
  const smokeApp = createApp();

  return async (req: Request): Promise<Response> => {
    return smokeApp.fetch(req, env);
  };
}

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
      configureServer(server) {
        server.middlewares.use(createDevBackendMiddleware(createSmokeBackendMiddleware()));
      },
    },
  ],
});
