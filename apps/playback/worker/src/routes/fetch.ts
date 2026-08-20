import { ConfigurationError, ValidationError } from "../../../contracts/index.ts";
import {
  createPlaybackControllers,
  PlaybackRuntimeConfigError,
  type PlaybackEnv,
} from "../composition/root.ts";
import { createAudioResponse } from "./audio-response.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";
import { matchPlaybackRoute } from "./match-playback-route.ts";

function mapRuntimeConfigErrorToExternal(error: unknown): unknown {
  if (error instanceof PlaybackRuntimeConfigError) {
    return new ConfigurationError("設定を確認できません", { cause: error });
  }
  return error;
}

/**
 * Playback worker の HTTP 入口。
 *
 * @require request は標準 Fetch の Request。env は Cloudflare Workers native secrets/vars
 * @ensure 成功時は 200。失敗 JSON は ErrorResponse（code のみ）。音声成功時は episodeAudioContentType の byte。
 *   env に Drive の OAuth 値と DRIVE_FOLDER_ID が揃う時は GoogleDriveEpisodeRepository を使う。
 *   設定不備は Composition Root の Error を HTTP boundary へ委譲する
 * @invariant Controller に unknown を渡す。HTTP status を分類しない
 */
export async function fetch(request: Request, env: PlaybackEnv): Promise<Response> {
  const requestId = crypto.randomUUID();
  try {
    const { listEpisodesController, getEpisodeController, getEpisodeAudioController } =
      createPlaybackControllers(env);
    const url = new URL(request.url);
    const matched = matchPlaybackRoute(request.method, url.pathname);
    switch (matched.kind) {
      case "list": {
        const input: unknown = {};
        const body = await listEpisodesController(input);
        return Response.json(body, { status: 200 });
      }
      case "get": {
        const input: unknown = { episodeId: matched.episodeId };
        const body = await getEpisodeController(input);
        return Response.json(body, { status: 200 });
      }
      case "audio": {
        const input: unknown = { episodeId: matched.episodeId };
        const bytes = await getEpisodeAudioController(input);
        return createAudioResponse(bytes);
      }
      case "unmatched":
        // why: 未一致 path を episode_not_found にすると、無い episode と無い route が同じ code になる
        throw new ValidationError("method または path が契約に無い");
      default: {
        const exhaustive: never = matched;
        void exhaustive;
        throw new ValidationError("method または path が契約に無い");
      }
    }
  } catch (error) {
    return createHttpErrorResponse(mapRuntimeConfigErrorToExternal(error), requestId);
  }
}
