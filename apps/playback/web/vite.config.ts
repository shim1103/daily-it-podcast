import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import {
  createFakeGetAudioUseCase,
  createFakeListEpisodesUseCase,
} from "../worker/src/controllers/fake-use-cases.ts";
import { createApp } from "../worker/src/routes/app.ts";

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
        const handle = createDummyBackendMiddleware();
        server.middlewares.use(async (req, res, next) => {
          if (!req.url?.startsWith("/episodes")) {
            next();
            return;
          }
          const response = await handle(
            new Request(new URL(req.url, "http://localhost"), { method: req.method }),
          );
          response.headers.forEach((value, key) => {
            res.setHeader(key, value);
          });
          const body = Buffer.from(await response.arrayBuffer());

          // why: dummy 音声は数十 MB の無音 WAV。browser の <audio> は Range 応答が無いと
          //   全長 buffer 完了まで再生・seek できず、実質「再生されない」ように見える。
          //   dev middleware だけ Range/HEAD をエミュレートして streaming・seek 可能にする
          //   （本番相当の Hono app は変更しない）。
          const rangeHeader = req.headers.range;
          if (res.statusCode === 200 && response.status === 200 && body.byteLength > 0) {
            res.setHeader("Accept-Ranges", "bytes");
            res.setHeader("Content-Length", String(body.byteLength));
          }
          if (req.method === "HEAD") {
            res.statusCode = response.status;
            res.end();
            return;
          }
          const rangeMatch = rangeHeader?.match(/^bytes=(\d*)-(\d*)$/);
          if (response.status === 200 && rangeMatch && body.byteLength > 0) {
            const start = rangeMatch[1] === "" ? 0 : Number(rangeMatch[1]);
            const end = rangeMatch[2] === "" ? body.byteLength - 1 : Number(rangeMatch[2]);
            if (start <= end && end < body.byteLength) {
              res.statusCode = 206;
              res.setHeader("Content-Range", `bytes ${start}-${end}/${body.byteLength}`);
              res.setHeader("Content-Length", String(end - start + 1));
              res.end(body.subarray(start, end + 1));
              return;
            }
            res.statusCode = 416;
            res.setHeader("Content-Range", `bytes */${body.byteLength}`);
            res.end();
            return;
          }
          res.statusCode = response.status;
          res.end(body);
        });
      },
    },
  ],
});
