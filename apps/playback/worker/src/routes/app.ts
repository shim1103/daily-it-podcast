import { Hono } from "hono";
import {
  ValidationError,
  episodeAudioRoutePath,
  episodeRoutePath,
  listEpisodesPath,
} from "../../../contracts/index.ts";
import { createPlaybackControllers, type PlaybackEnv } from "../composition/root.ts";
import { createAudioResponse } from "./audio-response.ts";
import { mapRuntimeConfigErrorToExternal } from "./fetch.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";

// why: Hono の AppType は method chain の戻り値に route が載る。mutation の app.get では
//   typeof app が空 schema のままになり、hc<AppType>() が unknown になる
export const app = new Hono<{ Bindings: PlaybackEnv }>()
  .get(listEpisodesPath, async (c) => {
    const { listEpisodesController } = createPlaybackControllers(c.env);
    const input: unknown = {};
    const body = await listEpisodesController(input);
    return Response.json(body, { status: 200 });
  })
  .get(episodeRoutePath, async (c) => {
    const { getEpisodeController } = createPlaybackControllers(c.env);
    const input: unknown = { episodeId: c.req.param("episodeId") };
    const body = await getEpisodeController(input);
    return Response.json(body, { status: 200 });
  })
  .get(episodeAudioRoutePath, async (c) => {
    const { getEpisodeAudioController } = createPlaybackControllers(c.env);
    const input: unknown = { episodeId: c.req.param("episodeId") };
    const bytes = await getEpisodeAudioController(input);
    return createAudioResponse(bytes);
  })
  .notFound(() => {
    // why: 未一致 path を episode_not_found にすると、無い episode と無い route が同じ code になる
    throw new ValidationError("method または path が契約に無い");
  })
  .onError((error) => {
    return createHttpErrorResponse(mapRuntimeConfigErrorToExternal(error), crypto.randomUUID());
  });

export type AppType = typeof app;
