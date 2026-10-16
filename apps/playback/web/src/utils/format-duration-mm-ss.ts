import type { DurationSec, PositionSec } from "../../../domain/index.ts";

/**
 * 長さまたは位置の秒（`DurationSec`・`PositionSec`）を mm:ss 表示へ変換する。
 *
 * @require 引数は 0 以上の秒数
 * @ensure ゼロ埋めの mm:ss（単位文字なし）を返す。分は最低 2 桁
 * @invariant 副作用なし。外部依存なし
 */
export function formatDurationMmSs(sec: DurationSec | PositionSec): string {
  const totalSec = Math.floor(sec);
  const minutes = Math.floor(totalSec / 60);
  const seconds = totalSec % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}
