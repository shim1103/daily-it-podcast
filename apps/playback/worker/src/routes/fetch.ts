import { UnavailableError, ValidationError } from "../../../contracts/index.ts";
import { createPlaybackControllers, type PlaybackEnv } from "../composition/root.ts";
import { createAudioResponse } from "./audio-response.ts";
import { createHttpErrorResponse } from "./http-error-response.ts";
import { matchPlaybackRoute } from "./match-playback-route.ts";

/**
 * Playback worker の HTTP 入口。
 *
 * @require request は標準 Fetch の Request。env は Cloudflare Workers native secrets/vars
 * @ensure 成功時は 200。失敗 JSON は ErrorResponse（code のみ）。音声成功時は episodeAudioContentType の byte。
 *   env に Drive の OAuth 値と DRIVE_FOLDER_ID が揃う時は GoogleDriveEpisodeRepository を使う。
 *   一部だけ設定されている（未設定と誤認できない設定漏れ）時は 503 unavailable を返す
 * @invariant Controller に unknown を渡す。classifyHttpStatus を呼ばない
 */
export async function fetch(request: Request, env: PlaybackEnv): Promise<Response> {
  const requestId = crypto.randomUUID();
  try {
    const controllersResult = createPlaybackControllers(env);
    if (controllersResult.kind === "misconfigured") {
      // why: OAuth 値の一部だけが欠けた状態は本番相当環境での設定漏れとみなす。
      // 無言で Fake へ落ちず、observable な失敗（503 + log）にする。
      throw new UnavailableError(
        `Drive 接続用 env が一部だけ欠落: ${controllersResult.missing.join(", ")}`,
      );
    }
    const { listEpisodesController, getEpisodeController, getEpisodeAudioController } =
      controllersResult.controllers;
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
    return createHttpErrorResponse(error, requestId);
  }
}
