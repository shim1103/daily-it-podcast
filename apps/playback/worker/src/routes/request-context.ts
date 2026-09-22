import type { MiddlewareHandler } from "hono";
import { logInfo, type RequestId } from "./logger.ts";

export type { RequestId } from "./logger.ts";

function createRequestId(): RequestId {
  return crypto.randomUUID() as RequestId;
}

export type RequestContextVariables = {
  requestId: RequestId;
};

export const requestIdHeaderName = "X-Request-Id";

/**
 * requestId の発行・伝播と、request 単位のアクセスログを担う Hono middleware。
 *
 * @require なし
 * @ensure system 境界に最初に入った地点で requestId を 1 度だけ発行し、`c.set("requestId", ...)` で
 *   後続 handler・onError へ伝播する。正常応答には requestId header を付与する。
 *   開始ログ（request_start）と完了ログ（request_end）を、成功・失敗どちらの経路でも各 1 回ずつ出す
 * @invariant log 出力はこの middleware と onError の 2 箇所に限定する。domain・infrastructure 層では出さない
 * @invariant requestId の発行はこの middleware に閉じる。他の層で crypto.randomUUID() を requestId
 *   として再発行しない
 */
export const requestLoggingMiddleware: MiddlewareHandler<{
  Variables: RequestContextVariables;
}> = async (c, next) => {
  const requestId = createRequestId();
  c.set("requestId", requestId);
  const method = c.req.method;
  const path = c.req.path;
  const start = Date.now();

  logInfo({ event: "request_start", requestId, method, path });

  await next();

  c.header(requestIdHeaderName, requestId);
  logInfo({
    event: "request_end",
    requestId,
    method,
    path,
    status: c.res.status,
    durationMs: Date.now() - start,
  });
};
