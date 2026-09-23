/**
 * 進捗 D1 永続の Infrastructure 定数。
 * Domain の完走ゾーンは `entities/constants/progress.ts` を正とする。
 */

/**
 * Workers D1 binding 名。`wrangler.jsonc` の `d1_databases[].binding` と
 * PlaybackEnv の key と一致させる（database 実体の作成・結線は C）。
 */
export const EPISODE_PROGRESS_D1_BINDING = "EPISODE_PROGRESS" as const;

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
