import type { MiddlewareHandler } from "hono";
import {
  createPlaybackControllers,
  type PlaybackControllers,
  type PlaybackEnv,
  type PlaybackUseCaseOverrides,
} from "../composition/root.ts";
import type { RequestContextVariables } from "./request-context.ts";

/**
 * route が読む request-scoped な Controller 一式。
 * Composition Root 自体は変えず、呼び口を pipeline に1回寄せる。
 */
export type PlaybackControllersVariables = RequestContextVariables & {
  controllers: PlaybackControllers;
};

/**
 * 1 request につき Composition Root を 1 回だけ呼び、Controller 一式を context へ載せる。
 *
 * cross-cutting（requestId / log / secureHeaders）とは別段: こちらは **request-scoped wiring**。
 *
 * @require 後段 route は `c.get("controllers")` だけを使い、個別に CR を呼ばない
 * @ensure `controllers` が Variables に載る。mode は production と同じ `"r2"` 固定
 * @invariant ビジネス判断・Port 直接呼び出しはしない（結線だけ）
 */
export function createPlaybackControllersMiddleware(
  useCaseOverrides?: PlaybackUseCaseOverrides,
): MiddlewareHandler<{
  Bindings: PlaybackEnv;
  Variables: PlaybackControllersVariables;
}> {
  return async (c, next) => {
    c.set("controllers", createPlaybackControllers(c.env, { mode: "r2" }, useCaseOverrides));
    await next();
  };
}
