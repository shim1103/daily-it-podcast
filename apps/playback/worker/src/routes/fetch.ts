import {
  UnavailableError,
  ValidationError,
  episodeAudioContentType,
  listEpisodesPath,
  type PlaybackHttpErrorCode,
} from "../../../contracts/index.ts";
import { createPlaybackControllers, type PlaybackEnv } from "../composition/root.ts";

type ExternalErrorName = "ValidationError" | "NotFoundError" | "UnavailableError";

type HttpErrorMapping = {
  readonly status: number;
  readonly code: PlaybackHttpErrorCode;
};

const externalHttpErrorMapping: {
  readonly [K in ExternalErrorName]: HttpErrorMapping;
} = {
  ValidationError: { status: 400, code: "validation_error" },
  NotFoundError: { status: 404, code: "episode_not_found" },
  UnavailableError: { status: 503, code: "unavailable" },
};

type MatchedRoute =
  | { kind: "list" }
  | { kind: "get"; episodeId: unknown }
  | { kind: "audio"; episodeId: unknown }
  | { kind: "unmatched" };

type CauseLog = {
  name: string;
  message: string;
  cause?: CauseLog;
};

type ErrorLogPayload = {
  name: string;
  message: string;
  stack: string | undefined;
  cause: CauseLog | undefined;
  requestId: string;
};

function toArrayBufferFromUint8Array(bytes: Uint8Array): ArrayBuffer {
  // why: `new Response(...)` の body に渡せるのは `BodyInit`。`Uint8Array` を型的に弾くため、
  // 返すべき可視範囲だけ `ArrayBuffer` にする。
  //
  // NOTE: `Uint8Array.buffer` は型的に `ArrayBuffer | SharedArrayBuffer` なので、
  // `SharedArrayBuffer` の場合は copy を作って確実に `ArrayBuffer` を返す。
  const buf = bytes.buffer;
  if (typeof SharedArrayBuffer !== "undefined" && buf instanceof SharedArrayBuffer) {
    const copy = new Uint8Array(bytes.byteLength);
    copy.set(bytes);
    return copy.buffer;
  }

  // ArrayBuffer の場合は slice で visible range を切り出す
  // NOTE: TS が union の slice 戻りを ArrayBuffer に絞り込めない可能性があるため、cast する。
  return (buf as ArrayBuffer).slice(
    bytes.byteOffset,
    bytes.byteOffset + bytes.byteLength,
  );
}

function isMappedExternalErrorName(name: string): name is ExternalErrorName {
  return Object.hasOwn(externalHttpErrorMapping, name);
}

function decodePathSegment(segment: string): string {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

function matchPlaybackRoute(method: string, pathname: string): MatchedRoute {
  if (method !== "GET") {
    return { kind: "unmatched" };
  }
  if (pathname === listEpisodesPath) {
    return { kind: "list" };
  }
  const prefix = `${listEpisodesPath}/`;
  if (!pathname.startsWith(prefix)) {
    return { kind: "unmatched" };
  }
  const rest = pathname.slice(prefix.length);
  const segments = rest.split("/");
  if (segments.length === 1) {
    return { kind: "get", episodeId: decodePathSegment(segments[0] ?? "") };
  }
  if (segments.length === 2 && segments[1] === "audio") {
    return { kind: "audio", episodeId: decodePathSegment(segments[0] ?? "") };
  }
  return { kind: "unmatched" };
}

function toCauseLog(cause: unknown): CauseLog | undefined {
  if (!(cause instanceof Error)) {
    return undefined;
  }
  const nested = toCauseLog(cause.cause);
  if (nested === undefined) {
    return { name: cause.name, message: cause.message };
  }
  return { name: cause.name, message: cause.message, cause: nested };
}

function toErrorLogPayload(error: Error, requestId: string): ErrorLogPayload {
  return {
    name: error.name,
    message: error.message,
    stack: error.stack,
    cause: toCauseLog(error.cause),
    requestId,
  };
}

function toHttpErrorResponse(error: unknown, requestId: string): Response {
  if (error instanceof Error && isMappedExternalErrorName(error.name)) {
    const mapped = externalHttpErrorMapping[error.name];
    console.error(toErrorLogPayload(error, requestId));
    return Response.json({ code: mapped.code }, { status: mapped.status });
  }

  if (error instanceof Error) {
    console.error({
      name: "UnmappedError",
      message: error.message,
      stack: error.stack,
      cause: toCauseLog(error.cause),
      requestId,
    });
  } else {
    console.error({
      name: "UnmappedError",
      message: String(error),
      requestId,
    });
  }
  // why: UnmappedError を JSON の code にすると契約 enum を破る
  return new Response(null, { status: 500 });
}

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
        return new Response(toArrayBufferFromUint8Array(bytes), {
          status: 200,
          headers: { "Content-Type": episodeAudioContentType },
        });
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
    return toHttpErrorResponse(error, requestId);
  }
}
