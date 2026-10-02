/**
 * 進捗 D1 永続の Infrastructure 定数。
 * Domain の完走ゾーンは `entities/constants/progress.ts` を正とする。
 */

import { getTableName } from "drizzle-orm";
import { episodeProgressTable } from "./schema.ts";

/**
 * Workers D1 binding 名。`wrangler.jsonc` の `d1_databases[].binding` と
 * PlaybackEnv の key と一致させる（database 実体の作成・credential 登録は A）。
 */
export const EPISODE_PROGRESS_D1_BINDING = "EPISODE_PROGRESS" as const;

/**
 * Worker→D1 の1操作あたりの最大試行回数（初回含む）。
 * 1 = adapter 内で再試行しない。一時失敗は Infrastructure Error → External unavailable とし、
 * browser の HTTP retry（contracts の PROGRESS_WRITE_MAX_ATTEMPTS）に任せる。
 */
export const PROGRESS_D1_ADAPTER_MAX_ATTEMPTS = 1 as const;

/** D1 進捗表名。`schema.ts` の表定義から導く（literal を複製しない）。 */
export const EPISODE_PROGRESS_TABLE = getTableName(episodeProgressTable);

/** D1 進捗表の列名。`schema.ts` の列定義から導く（literal を複製しない）。 */
export const episodeProgressColumns = {
  episodeId: episodeProgressTable.episodeId.name,
  positionSec: episodeProgressTable.positionSec.name,
  firstPlayedAt: episodeProgressTable.firstPlayedAt.name,
  firstCompletedAt: episodeProgressTable.firstCompletedAt.name,
  lastPlayedAt: episodeProgressTable.lastPlayedAt.name,
  seq: episodeProgressTable.seq.name,
} as const;

/**
 * D1 の 1 query あたりの bind 値の上限個数。
 * what: IN 句の一括取得はこの個数で分割する。local D1 は 101 個目の bind で `too many SQL variables` を返す。
 */
export const D1_MAX_BOUND_PARAMETERS_PER_QUERY = 100 as const;
