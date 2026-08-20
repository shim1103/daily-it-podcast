import type { PlaybackApiErrorCode } from "./playback-api-error.ts";

export type ApiResult<T> = { ok: true; data: T } | { ok: false; error: PlaybackApiErrorCode };
