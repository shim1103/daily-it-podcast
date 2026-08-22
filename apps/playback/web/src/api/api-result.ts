import type { PlaybackApiErrorCode } from "./playback-api-error.ts";

export type ApiResult<T> = { ok: true; data: T } | { ok: false; error: PlaybackApiErrorCode };

/**
 * `ApiResult` を返す async method の型から、成功時 data の型を取り出す。
 *
 * why: ViewModel 層は境界共有型（契約 module）を直接importせず、API Client の型からのみ
 *   成功 data の型を導出する（importルール: ViewModel は境界共有型を API Client 経由で使う）
 */
export type ApiSuccessData<Method extends (...args: never[]) => Promise<ApiResult<unknown>>> =
  Extract<Awaited<ReturnType<Method>>, { ok: true }>["data"];
