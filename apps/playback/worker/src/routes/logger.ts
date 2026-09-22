/**
 * console.log/console.error を直接呼ぶ箇所をこの file に限定するための wrapper。
 * biome の `suspicious.noConsole` を全体で有効化し、この file だけ例外にする。
 */
type InfoPayload = {
  event: string;
  [key: string]: unknown;
};

/** @ensure console.log へ payload をそのまま渡す */
export function logInfo(payload: InfoPayload): void {
  console.log(payload);
}

/** @ensure console.error へ payload をそのまま渡す */
export function logError(payload: Record<string, unknown>): void {
  console.error(payload);
}
