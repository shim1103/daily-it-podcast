import { Hono } from "hono";
import {
  ValidationError,
  episodeAudioRoutePath,
  listEpisodesPath,
} from "../../../contracts/index.ts";
import {
  createPlaybackControllers,
  type PlaybackEnv,
  type PlaybackUseCaseOverrides,
} from "../composition/root.ts";
import { createAudioResponse } from "./audio-response.ts";
import { episodeListCacheHeaders } from "./cache-policy.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";
import { mapRuntimeConfigErrorToExternal } from "./runtime-config-error-mapping.ts";

/**
 * Playback worker の Hono instance を組み立てる。
 *
 * @require なし
 * @ensure 各 route は options として { mode: "r2" } を固定で Composition Root へ渡す。
 *   useCaseOverrides を渡す時はその override も併せてそのまま渡す
 * @invariant route 定義・Error 写像は production 用の `app` と同一のまま複製しない
 */
export function createApp(useCaseOverrides?: PlaybackUseCaseOverrides) {
  // why: Hono の AppType は method chain の戻り値に route が載る。mutation の instance.get では
  //   typeof app が空 schema のままになり、hc<AppType>() が unknown になる
  return new Hono<{ Bindings: PlaybackEnv }>()
    .get(listEpisodesPath, async (c) => {
      const { listEpisodesController } = createPlaybackControllers(
        c.env,
        { mode: "r2" },
        useCaseOverrides,
      );
      const input: unknown = {};
      const body = await listEpisodesController(input);
      return Response.json(body, { status: 200, headers: episodeListCacheHeaders });
    })
    .get(episodeAudioRoutePath, async (c) => {
      const { getAudioController } = createPlaybackControllers(
        c.env,
        { mode: "r2" },
        useCaseOverrides,
      );
      const input: unknown = { episodeId: c.req.param("episodeId") };
      const bytes = await getAudioController(input);
      return createAudioResponse(bytes, c.req.header("Range") ?? null);
    })
    .notFound(() => {
      // why: 未一致 path を episode_not_found にすると、無い episode と無い route が同じ code になる
      throw new ValidationError("method または path が契約に無い");
    })
    .onError((error) => {
      return createHttpErrorResponse(mapRuntimeConfigErrorToExternal(error), crypto.randomUUID());
    });
}

export const app = createApp();

export type AppType = typeof app;
