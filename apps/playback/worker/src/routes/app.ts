import { zValidator } from "@hono/zod-validator";
import { Hono } from "hono";
import { etag } from "hono/etag";
import { requestId } from "hono/request-id";
import { secureHeaders } from "hono/secure-headers";
import {
  EpisodeIdRequestSchema,
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
import {
  createRequestId,
  requestIdHeaderName,
  requestLoggingMiddleware,
  type RequestContextVariables,
} from "./request-context.ts";
import { mapRuntimeConfigErrorToExternal } from "./runtime-config-error-mapping.ts";

/**
 * zValidator("param", EpisodeIdRequestSchema) の失敗を ValidationError へ変換する。
 *
 * @require result は EpisodeIdRequestSchema の safeParse 結果
 * @ensure 不適合時は ValidationError を throw し、元の zod error を cause へ残す。適合時は何もしない
 */
export function throwOnEpisodeIdValidationFailure(result: { success: boolean; error?: unknown }) {
  if (!result.success) {
    throw new ValidationError("入力が契約に不適合", { cause: result.error });
  }
}

/**
 * Playback worker の Hono instance を組み立てる。
 *
 * @require なし
 * @ensure 各 route は options として { mode: "r2" } を固定で Composition Root へ渡す。
 *   useCaseOverrides を渡す時はその override も併せてそのまま渡す
 * @invariant route 定義・Error 写像は production 用の `app` と同一のまま複製しない
 */
export function createApp(useCaseOverrides?: PlaybackUseCaseOverrides) {
  return new Hono<{ Bindings: PlaybackEnv; Variables: RequestContextVariables }>()
    .use(requestId({ generator: createRequestId, headerName: requestIdHeaderName }))
    .use(requestLoggingMiddleware)
    .use(secureHeaders())
    .get(
      listEpisodesPath,
      // why: 音声GETには導入しない。判断根拠は
      //   docs/decisions/2026-09-23T03-44-58-feature-playback-etag-list-episodes.md
      etag(),
      async (c) => {
        const { listEpisodesController } = createPlaybackControllers(
          c.env,
          { mode: "r2" },
          useCaseOverrides,
        );
        const body = await listEpisodesController();
        return c.json(body, 200, episodeListCacheHeaders);
      },
    )
    .get(
      episodeAudioRoutePath,
      zValidator("param", EpisodeIdRequestSchema, throwOnEpisodeIdValidationFailure),
      async (c) => {
        const { getAudioController } = createPlaybackControllers(
          c.env,
          { mode: "r2" },
          useCaseOverrides,
        );
        const { episodeId } = c.req.valid("param");
        const bytes = await getAudioController(episodeId);
        return createAudioResponse(bytes, c.req.header("Range") ?? null);
      },
    )
    .notFound(() => {
      throw new ValidationError("method または path が契約に無い");
    })
    .onError((error, c) => {
      return createHttpErrorResponse(mapRuntimeConfigErrorToExternal(error), c.get("requestId"));
    });
}

export const app = createApp();

export type AppType = typeof app;
