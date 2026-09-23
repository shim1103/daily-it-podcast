/**
 * 再生進捗の Domain 定数（業務規則として固定する値だけ）。
 * D1 binding・表列・adapter 再試行は Infrastructure（`infrastructure/d1/`）を正とする。
 * browser→Worker の HTTP retry 上限は contracts を正とする。
 */

/** 完走ゾーン幅（秒）。`positionSec >= durationSec - この値` で complete 契機。 */
export const PROGRESS_COMPLETE_ZONE_SEC = 3 as const;
