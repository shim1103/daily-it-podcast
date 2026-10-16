/** 音声内の再生位置（秒）。非負であることは、入口の契約（zod の `min(0)`）が保証する。 */
export type PositionSec = number;
