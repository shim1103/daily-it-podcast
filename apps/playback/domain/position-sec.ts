/** 音声内の位置（秒）。再生位置（`positionSec`）と、原稿の各区間の開始位置（`startSec`）を表す。非負であることは、入口の契約（zod の `min(0)`）が保証する。 */
export type PositionSec = number;
