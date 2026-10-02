/**
 * 再生進捗の Domain 定数（業務規則として固定する値だけ）。
 * D1 binding・表列・adapter 再試行は Infrastructure（`infrastructure/d1/`）を正とする。
 * browser→Worker の HTTP retry 上限は contracts を正とする。
 */

/** 完走ゾーン幅（秒）。`positionSec >= durationSec - この値` で complete 契機。 */
export const PROGRESS_COMPLETE_ZONE_SEC = 3 as const;

// what: `clientAt` が `serverAt` より未来へずれてよい幅（秒）。
// why: 値の正本はこの定数（Decision 2026-10-01T23-34-00 で決めた仮値）。未来側は、通した時の被害の上界が上限と同じになるため狭く取る。
export const PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC = 300 as const;

// what: `clientAt` が `serverAt` より過去へずれてよい幅（秒）。
// why: 値の正本はこの定数（Decision 2026-10-01T23-34-00 で決めた仮値）。first* は先勝ちで、過去のずれが恒久に残って汚染するため、上限を置く。
export const PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC = 24 * 60 * 60;
