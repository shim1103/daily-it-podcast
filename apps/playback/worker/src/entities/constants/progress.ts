/**
 * 再生進捗の Domain 定数。D1 列契約と完走ゾーン・adapter 再試行の正本の一部。
 * browser→Worker の HTTP retry 上限は contracts を正とする。
 */

/** 完走ゾーン幅（秒）。`positionSec >= durationSec - この値` で complete 契機。 */
export const PROGRESS_COMPLETE_ZONE_SEC = 3 as const;

/**
 * Worker→D1 の1操作あたりの最大試行回数（初回含む）。
 * 1 = adapter 内で再試行しない。一時失敗は Infrastructure Error → External unavailable とし、
 * browser の HTTP retry（contracts の PROGRESS_WRITE_MAX_ATTEMPTS）に任せる。
 */
export const PROGRESS_D1_ADAPTER_MAX_ATTEMPTS = 1 as const;

/** D1 進捗表名。migration と adapter が共有する。 */
export const EPISODE_PROGRESS_TABLE = "episode_progress" as const;

/** D1 進捗表の列名。migration と adapter が共有する。 */
export const episodeProgressColumns = {
  episodeId: "episode_id",
  positionSec: "position_sec",
  firstPlayedAt: "first_played_at",
  firstCompletedAt: "first_completed_at",
  lastPlayedAt: "last_played_at",
} as const;
