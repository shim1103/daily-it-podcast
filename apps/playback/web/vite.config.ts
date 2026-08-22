import { defineConfig } from "vite";
import {
  createFakeGetEpisodeAudioUseCase,
  createFakeGetEpisodeUseCase,
  createFakeListEpisodesUseCase,
} from "../worker/src/controllers/fake-use-cases.ts";
import { createPlaybackControllers } from "../worker/src/composition/root.ts";
import { createAudioResponse } from "../worker/src/routes/audio-response.ts";
import { createHttpErrorResponse } from "../worker/src/routes/http-error-response.ts";
import { matchPlaybackRoute } from "../worker/src/routes/match-playback-route.ts";

/**
 * localhost:3000 単体起動で dummy backend（fake-use-cases）由来の data を流す dev-only middleware。
 *
 * @require なし（fake-use-cases 由来の use case を Composition Root の overrides 経由で固定する）
 * @ensure matchPlaybackRoute の一致に応じて Controller を呼び分け、Response を返す。
 *   未一致は undefined を返し、呼び出し側の next handler へ委ねる
 * @invariant 本番相当の HTTP entry は worker/src/routes/fetch.ts のまま変更しない
 */
export function createDummyBackendMiddleware() {
  const { listEpisodesController, getEpisodeController, getEpisodeAudioController } =
    createPlaybackControllers(
      {},
      {},
      {
        useCases: {
          listEpisodes: createFakeListEpisodesUseCase(),
          getEpisode: createFakeGetEpisodeUseCase(),
          getEpisodeAudio: createFakeGetEpisodeAudioUseCase(),
        },
      },
    );

  return async (req: Request): Promise<Response | undefined> => {
    const url = new URL(req.url, "http://localhost");
    const matched = matchPlaybackRoute(req.method, url.pathname);
    const requestId = crypto.randomUUID();
    try {
      switch (matched.kind) {
        case "list":
          return Response.json(await listEpisodesController({}), { status: 200 });
        case "get":
          return Response.json(await getEpisodeController({ episodeId: matched.episodeId }), {
            status: 200,
          });
        case "audio":
          return createAudioResponse(
            await getEpisodeAudioController({ episodeId: matched.episodeId }),
          );
        case "unmatched":
          return undefined;
        default:
          return undefined;
      }
    } catch (error) {
      return createHttpErrorResponse(error, requestId);
    }
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
  plugins: [
    {
      name: "dummy-backend-api",
      configureServer(server) {
        const handle = createDummyBackendMiddleware();
        server.middlewares.use(async (req, res, next) => {
          if (!req.url?.startsWith("/episodes")) {
            next();
            return;
          }
          const response = await handle(
            new Request(new URL(req.url, "http://localhost"), { method: req.method }),
          );
          if (!response) {
            next();
            return;
          }
          res.statusCode = response.status;
          response.headers.forEach((value, key) => {
            res.setHeader(key, value);
          });
          const body = Buffer.from(await response.arrayBuffer());
          res.end(body);
        });
      },
    },
  ],
});
