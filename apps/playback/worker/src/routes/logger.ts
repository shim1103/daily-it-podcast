import type { RequestId } from "./request-context.ts";

/**
 * console.log/console.error を直接呼ぶ箇所をこの file に限定するための wrapper。
 * biome の `suspicious.noConsole` を全体で有効化し、この file だけ例外にする。
 */
type InfoPayload = {
  event: string;
  requestId?: RequestId;
  [key: string]: unknown;
};

/** @ensure console.log へ payload をそのまま渡す */
export function logInfo(payload: InfoPayload): void {
  console.log(payload);
}

/**
 * Error 診断 payload（`{ name, message, stack, cause, requestId }` 形）を渡す。
 * `logInfo` とは形が異なる（`event` を持たない）ため型を分ける。
 */
export function logError(payload: Record<string, unknown>): void {
  console.error(payload);
}
