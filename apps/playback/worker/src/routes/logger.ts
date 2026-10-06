/**
 * console.log/console.error を直接呼ぶ箇所をこの file に限定するための wrapper。
 * biome の `suspicious.noConsole` を全体で有効化し、この file だけ例外にする。
 */

/**
 * system 境界（requestLoggingMiddleware）でのみ発行する branded string。
 * ただの `string` と型上区別し、任意の文字列を requestId として誤って渡せないようにする。
 * 発行元は `request-context.ts` だが、log payload が requestId を必須で持つことを
 * この file（logger.ts）が強制するため、型定義はここに置く。
 */
export type RequestId = string & { readonly __brand: "RequestId" };

type InfoPayload = {
  event: string;
  requestId: RequestId;
  [key: string]: unknown;
};

type ErrorPayload = {
  name: string;
  message: string;
  requestId: RequestId;
  [key: string]: unknown;
};

/**
 * @require payload は event と requestId を持つ
 * @ensure console.log へ payload をそのまま渡す
 */
export function logInfo(payload: InfoPayload): void {
  console.log(payload);
}

/**
 * @require payload は name・message・requestId を持つ
 * @ensure console.error へ payload をそのまま渡す
 */
export function logError(payload: ErrorPayload): void {
  console.error(payload);
}

/**
 * request に紐づかない出力（CLI の標準出力など）用。structured log ではなく 1 行の文字列を出す。
 * @ensure console.log へ 1 行をそのまま渡す
 */
export function writeLine(line: string): void {
  console.log(line);
}

/**
 * request に紐づかない出力（CLI の標準エラーなど）用。
 * @ensure console.error へ 1 行をそのまま渡す
 */
export function writeErrorLine(line: string): void {
  console.error(line);
}
