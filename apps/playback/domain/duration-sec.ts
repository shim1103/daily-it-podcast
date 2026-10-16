/** 音声全体の長さ（秒）。非負であることは、入口の契約（zod の `min(0)`）と JSON Schema が保証する。 */
export type DurationSec = number;
