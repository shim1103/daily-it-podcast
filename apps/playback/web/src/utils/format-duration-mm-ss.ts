/**
 * durationSec を mm:ss 表示へ変換する。
 *
 * @require durationSec は 0 以上の秒数
 * @ensure ゼロ埋めの mm:ss（単位文字なし）を返す。分は最低 2 桁
 * @invariant 副作用なし。外部依存なし
 */
export function formatDurationMmSs(durationSec: number): string {
  const totalSec = Math.floor(durationSec);
  const minutes = Math.floor(totalSec / 60);
  const seconds = totalSec % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}
