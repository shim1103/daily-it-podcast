import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";
import {
  EPISODE_PROGRESS_TABLE,
  episodeProgressColumns,
} from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";

export type ProgressProbeRow = {
  episodeId: string;
  positionSec: number;
  firstPlayedAt: string;
  firstCompletedAt: string | null;
  lastPlayedAt: string;
};

const column = episodeProgressColumns;

const insertSql = `INSERT INTO ${EPISODE_PROGRESS_TABLE} (${column.episodeId}, ${column.positionSec}, ${column.firstPlayedAt}, ${column.firstCompletedAt}, ${column.lastPlayedAt}) VALUES (?, ?, ?, ?, ?)`;

const selectSql = `SELECT ${column.episodeId} AS episodeId, ${column.positionSec} AS positionSec, ${column.firstPlayedAt} AS firstPlayedAt, ${column.firstCompletedAt} AS firstCompletedAt, ${column.lastPlayedAt} AS lastPlayedAt FROM ${EPISODE_PROGRESS_TABLE} WHERE ${column.episodeId} = ?`;

const deleteSql = `DELETE FROM ${EPISODE_PROGRESS_TABLE} WHERE ${column.episodeId} = ?`;

export function createProbeRow(episodeId: string): ProgressProbeRow {
  return {
    episodeId,
    positionSec: 12.5,
    firstPlayedAt: "2026-09-30T00:00:00.000Z",
    firstCompletedAt: null,
    lastPlayedAt: "2026-09-30T00:00:12.500Z",
  };
}

/** 書けた行数を返す。 */
export async function writeProbeRow(
  database: D1DatabaseBinding,
  row: ProgressProbeRow,
): Promise<number> {
  const written = await database
    .prepare(insertSql)
    .bind(row.episodeId, row.positionSec, row.firstPlayedAt, row.firstCompletedAt, row.lastPlayedAt)
    .run();
  return written.meta.changes;
}

/** 行が無ければ null を返す。 */
export async function readProbeRow(
  database: D1DatabaseBinding,
  episodeId: string,
): Promise<ProgressProbeRow | null> {
  return database.prepare(selectSql).bind(episodeId).first<ProgressProbeRow>();
}

/** 消せた行数を返す。 */
export async function deleteProbeRow(
  database: D1DatabaseBinding,
  episodeId: string,
): Promise<number> {
  const deleted = await database.prepare(deleteSql).bind(episodeId).run();
  return deleted.meta.changes;
}
