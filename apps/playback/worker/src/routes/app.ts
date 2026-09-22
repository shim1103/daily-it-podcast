import { zValidator } from "@hono/zod-validator";
import { Hono } from "hono";
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
import { requestLoggingMiddleware, type RequestContextVariables } from "./request-context.ts";
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
  // why: Hono の AppType は method chain の戻り値に route が載る。mutation の instance.get では
  //   typeof app が空 schema のままになり、hc<AppType>() が unknown になる
  return new Hono<{ Bindings: PlaybackEnv; Variables: RequestContextVariables }>()
    .use(requestLoggingMiddleware)
    .get(listEpisodesPath, async (c) => {
      const { listEpisodesController } = createPlaybackControllers(
        c.env,
        { mode: "r2" },
        useCaseOverrides,
      );
      const input: unknown = {};
      const body = await listEpisodesController(input);
      return c.json(body, 200, episodeListCacheHeaders);
    })
    .get(
      episodeAudioRoutePath,
      // why: zValidator は HTTP 入口での早期検証。GetAudioController は unknown を受ける契約を
      //   保つため、controller 内の parseEpisodeIdRequest による再検証はそのまま残す（二重検証）。
      //   controller が route（Hono）に依存せず単独で安全なまま再利用・test できることを優先する
      zValidator("param", EpisodeIdRequestSchema, throwOnEpisodeIdValidationFailure),
      async (c) => {
        const { getAudioController } = createPlaybackControllers(
          c.env,
          { mode: "r2" },
          useCaseOverrides,
        );
        const input: unknown = c.req.valid("param");
        const bytes = await getAudioController(input);
        return createAudioResponse(bytes, c.req.header("Range") ?? null);
      },
    )
    .notFound(() => {
      // why: 未一致 path を episode_not_found にすると、無い episode と無い route が同じ code になる
      throw new ValidationError("method または path が契約に無い");
    })
    .onError((error, c) => {
      // why: requestId は境界（requestLoggingMiddleware）で 1 度だけ発行した値をそのまま使う。
      //   ここで再発行すると、開始・完了ログと error ログの requestId が食い違う
      return createHttpErrorResponse(mapRuntimeConfigErrorToExternal(error), c.get("requestId"));
    });
}

export const app = createApp();

export type AppType = typeof app;
