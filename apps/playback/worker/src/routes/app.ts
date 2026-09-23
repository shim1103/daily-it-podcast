import { zValidator } from "@hono/zod-validator";
import { Hono } from "hono";
import { etag } from "hono/etag";
import { requestId } from "hono/request-id";
import { secureHeaders } from "hono/secure-headers";
import {
  EpisodeIdRequestSchema,
  ProgressPullQuerySchema,
  ProgressWriteRequestSchema,
  ValidationError,
  episodeAudioRoutePath,
  episodeProgressCompleteRoutePath,
  episodeProgressRoutePath,
  listEpisodesPath,
  progressPullPath,
} from "../../../contracts/index.ts";
import type { PlaybackEnv, PlaybackUseCaseOverrides } from "../composition/root.ts";
import { createAudioResponse } from "./audio-response.ts";
import {
  episodeListCacheHeaders,
  progressPullCacheHeaders,
  progressWriteCacheHeaders,
} from "./cache-policy.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";
import {
  createPlaybackControllersMiddleware,
  type PlaybackControllersVariables,
} from "./playback-controllers-middleware.ts";
import {
  createRequestId,
  requestIdHeaderName,
  requestLoggingMiddleware,
} from "./request-context.ts";
import { mapRuntimeConfigErrorToExternal } from "./runtime-config-error-mapping.ts";

/**
 * zValidator の失敗を ValidationError へ変換する。
 *
 * @require result は契約 schema の safeParse 結果
 * @ensure 不適合時は ValidationError を throw し、元の zod error を cause へ残す。適合時は何もしない
 */
export function throwOnContractValidationFailure(result: { success: boolean; error?: unknown }) {
  if (!result.success) {
    throw new ValidationError("入力が契約に不適合", { cause: result.error });
  }
}

/**
 * Playback worker の Hono instance を組み立てる。
 *
 * @require なし
 * @ensure Controllers は request-scoped middleware が 1 回だけ Composition Root を呼んで載せる。
 *   useCaseOverrides を渡す時はその override も middleware へ渡す
 * @invariant route 定義・Error 写像は production 用の `app` と同一のまま複製しない
 */
export function createApp(useCaseOverrides?: PlaybackUseCaseOverrides) {
  return new Hono<{ Bindings: PlaybackEnv; Variables: PlaybackControllersVariables }>()
    .use(requestId({ generator: createRequestId, headerName: requestIdHeaderName }))
    .use(requestLoggingMiddleware)
    .use(secureHeaders())
    .use(createPlaybackControllersMiddleware(useCaseOverrides))
    .get(
      listEpisodesPath,
      // why: 音声GETには導入しない。判断根拠は
      //   docs/decisions/2026-09-23T03-44-58-feature-playback-etag-list-episodes.md
      etag(),
      async (c) => {
        const { listEpisodesController } = c.get("controllers");
        const body = await listEpisodesController();
        return c.json(body, 200, episodeListCacheHeaders);
      },
    )
    .get(
      episodeAudioRoutePath,
      zValidator("param", EpisodeIdRequestSchema, throwOnContractValidationFailure),
      async (c) => {
        const { getAudioController } = c.get("controllers");
        const { episodeId } = c.req.valid("param");
        const bytes = await getAudioController(episodeId);
        return createAudioResponse(bytes, c.req.header("Range") ?? null);
      },
    )
    .post(
      episodeProgressRoutePath,
      zValidator("param", EpisodeIdRequestSchema, throwOnContractValidationFailure),
      zValidator("json", ProgressWriteRequestSchema, throwOnContractValidationFailure),
      async (c) => {
        const { createProgressController } = c.get("controllers");
        const { episodeId } = c.req.valid("param");
        const body = await createProgressController(episodeId, c.req.valid("json"));
        return c.json(body, 200, progressWriteCacheHeaders);
      },
    )
    .patch(
      episodeProgressRoutePath,
      zValidator("param", EpisodeIdRequestSchema, throwOnContractValidationFailure),
      zValidator("json", ProgressWriteRequestSchema, throwOnContractValidationFailure),
      async (c) => {
        const { updateProgressController } = c.get("controllers");
        const { episodeId } = c.req.valid("param");
        const body = await updateProgressController(episodeId, c.req.valid("json"));
        return c.json(body, 200, progressWriteCacheHeaders);
      },
    )
    .post(
      episodeProgressCompleteRoutePath,
      zValidator("param", EpisodeIdRequestSchema, throwOnContractValidationFailure),
      zValidator("json", ProgressWriteRequestSchema, throwOnContractValidationFailure),
      async (c) => {
        const { completeProgressController } = c.get("controllers");
        const { episodeId } = c.req.valid("param");
        const body = await completeProgressController(episodeId, c.req.valid("json"));
        return c.json(body, 200, progressWriteCacheHeaders);
      },
    )
    .get(
      progressPullPath,
      zValidator("query", ProgressPullQuerySchema, throwOnContractValidationFailure),
      async (c) => {
        const { pullProgressController } = c.get("controllers");
        const { since } = c.req.valid("query");
        const body = await pullProgressController(since);
        return c.json(body, 200, progressPullCacheHeaders);
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
