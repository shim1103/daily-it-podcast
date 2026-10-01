import { PROGRESS_COMPLETE_ZONE_SEC } from "../../entities/constants/progress.ts";

/**
 * 再生位置が完走ゾーンに入っているかを判定する。
 * 短尺（`durationSec` がゾーン幅未満）も同じ不等式（Decision 2026-10-01T23-09-47）。
 *
 * @require `positionSec` / `durationSec` は非負。`durationSec` は原稿正本の全長
 * @ensure `positionSec >= durationSec - PROGRESS_COMPLETE_ZONE_SEC` の時だけ true
 */
export function isInCompleteZone(positionSec: number, durationSec: number): boolean {
  return positionSec >= durationSec - PROGRESS_COMPLETE_ZONE_SEC;
}
