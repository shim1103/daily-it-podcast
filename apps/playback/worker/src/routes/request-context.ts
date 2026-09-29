import type { MiddlewareHandler } from "hono";
import { logInfo, type RequestId } from "./logger.ts";

export type { RequestId } from "./logger.ts";

/**
 * `hono/request-id` の `generator` option に渡す。requestId を branded type として発行する。
 *
 * @ensure crypto.randomUUID() を RequestId へ cast して返す
 */
export function createRequestId(): RequestId {
  return crypto.randomUUID() as RequestId;
}

export type RequestContextVariables = {
  requestId: RequestId;
};

export const requestIdHeaderName = "X-Request-Id";

/**
 * request 単位のアクセスログを担う Hono middleware。
 *
 * @require `hono/request-id`（`requestId({ generator: createRequestId })`）が、この middleware
 *   より前段で requestId を発行・`c.set` 済みであること
 * @ensure 開始ログ（request_start）と完了ログ（request_end）を、成功・失敗どちらの経路でも
 *   各 1 回ずつ出す
 * @invariant log 出力はこの middleware と onError の 2 箇所に限定する。domain・infrastructure 層では出さない
 * @invariant requestId の発行はこの middleware で行わない。`hono/request-id` 側の発行結果を読むだけ
 */
export const requestLoggingMiddleware: MiddlewareHandler<{
  Variables: RequestContextVariables;
}> = async (c, next) => {
  const requestId = c.get("requestId");
  const method = c.req.method;
  const path = c.req.path;
  const start = Date.now();

  logInfo({ event: "request_start", requestId, method, path });

  await next();

  logInfo({
    event: "request_end",
    requestId,
    method,
    path,
    status: c.res.status,
    durationMs: Date.now() - start,
  });
};
