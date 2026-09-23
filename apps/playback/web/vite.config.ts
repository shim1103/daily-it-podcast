import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import {
  createFakeGetAudioUseCase,
  createFakeListEpisodesUseCase,
  createFakeProgressWriteUseCase,
  createFakePullProgressUseCase,
} from "../worker/src/controllers/fake-use-cases.ts";
import { createApp } from "../worker/src/routes/app.ts";
import { createDevBackendMiddleware } from "./dev-backend-proxy.ts";

/**
 * localhost:3000 単体起動で dummy backend（fake-use-cases）由来の data を流す dev-only middleware。
 *
 * @require なし（fake-use-cases 由来の use case を Composition Root の overrides 経由で固定する）
 * @ensure worker/src/routes/app.ts の Hono instance へ fake use case override を注入して呼ぶ。
 *   route 未一致は app.ts の notFound handler が ValidationError を throw し、
 *   onError が 400 validation_error へ変換した応答を返す（Hono の素の 404 応答ではない）
 * @invariant 本番相当の HTTP entry は worker/src/routes/app.ts のまま変更しない
 */
export function createDummyBackendMiddleware() {
  const dummyApp = createApp({
    useCases: {
      listEpisodes: createFakeListEpisodesUseCase(),
      getAudio: createFakeGetAudioUseCase(),
      createProgress: createFakeProgressWriteUseCase(),
      updateProgress: createFakeProgressWriteUseCase(),
      completeProgress: createFakeProgressWriteUseCase(),
      pullProgress: createFakePullProgressUseCase(),
    },
  });

  return async (req: Request): Promise<Response> => {
    return dummyApp.fetch(req, {});
  };
}

export default defineConfig({
  // why: `apps/playback` から `--config web/vite.config.ts` で起動する運用のため、
  //   root を config file の場所へ明示する（未指定だと root が process.cwd() になり index.html を解決できない）
  root: new URL(".", import.meta.url).pathname,
  server: {
    port: 3000,
    middlewareMode: false,
  },
  build: {
    // why: wrangler.jsonc の assets.directory（./web/dist）と一致させる。root が web/ なので outDir は dist のみ
    outDir: "dist",
    emptyOutDir: true,
  },
  plugins: [
    react(),
    {
      name: "dummy-backend-api",
      configureServer(server) {
        // why: Range/HEAD は Hono app 自身（routes/audio-response.ts）が処理する。
        //   dev-backend-proxy はそれを透過中継するだけで streaming・seek が成立する。
        server.middlewares.use(createDevBackendMiddleware(createDummyBackendMiddleware()));
      },
    },
  ],
});
