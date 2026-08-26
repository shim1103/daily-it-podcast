import { Hono } from "hono";
import { listEpisodesPath, ValidationError } from "../../../contracts/index.ts";
import {
  createPlaybackControllers,
  type PlaybackEnv,
  type PlaybackUseCaseOverrides,
} from "../composition/root.ts";
import { createAudioResponse } from "./audio-response.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";
import { mapRuntimeConfigErrorToExternal } from "./runtime-config-error-mapping.ts";

/**
 * Playback worker の Hono instance を組み立てる。
 *
 * @require なし
 * @ensure useCaseOverrides を渡す時は各 route が Composition Root へそのまま渡す。
 *   省略時は production 相当（env のみ）で Controller を組み立てる
 * @invariant route 定義・Error 写像は production 用の `app` と同一のまま複製しない
 */
export function createApp(useCaseOverrides?: PlaybackUseCaseOverrides) {
  const instance = new Hono<{ Bindings: PlaybackEnv }>();

  instance.get(listEpisodesPath, async (c) => {
    const { listEpisodesController } = createPlaybackControllers(
      c.env,
      undefined,
      useCaseOverrides,
    );
    const input: unknown = {};
    const body = await listEpisodesController(input);
    return Response.json(body, { status: 200 });
  });

  instance.get(`${listEpisodesPath}/:episodeId`, async (c) => {
    const { getEpisodeController } = createPlaybackControllers(c.env, undefined, useCaseOverrides);
    const input: unknown = { episodeId: c.req.param("episodeId") };
    const body = await getEpisodeController(input);
    return Response.json(body, { status: 200 });
  });

  instance.get(`${listEpisodesPath}/:episodeId/audio`, async (c) => {
    const { getEpisodeAudioController } = createPlaybackControllers(
      c.env,
      undefined,
      useCaseOverrides,
    );
    const input: unknown = { episodeId: c.req.param("episodeId") };
    const bytes = await getEpisodeAudioController(input);
    return createAudioResponse(bytes);
  });

  instance.notFound(() => {
    // why: 未一致 path を episode_not_found にすると、無い episode と無い route が同じ code になる
    throw new ValidationError("method または path が契約に無い");
  });

  instance.onError((error) => {
    return createHttpErrorResponse(mapRuntimeConfigErrorToExternal(error), crypto.randomUUID());
  });

  return instance;
}

export const app = createApp();

export type AppType = typeof app;
