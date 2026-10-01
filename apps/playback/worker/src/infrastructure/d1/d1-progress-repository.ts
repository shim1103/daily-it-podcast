import type {
  ProgressRepository,
  ProgressUpdatedEntry,
  ProgressWriteCommand,
} from "../../application/ports/progress-repository.ts";
import type { D1DatabaseBinding, D1Row } from "./d1-database-binding.ts";
import { D1Error } from "./d1-error.ts";
import {
  D1_MAX_BOUND_PARAMETERS_PER_QUERY,
  EPISODE_PROGRESS_TABLE,
  episodeProgressColumns,
} from "./progress-d1-constants.ts";

type EpisodeProgress = ProgressUpdatedEntry["progress"];
type ProgressWriteResult = Awaited<ReturnType<ProgressRepository["writeProgress"]>>;

const columns = episodeProgressColumns;
const TABLE = EPISODE_PROGRESS_TABLE;

const ALL_COLUMNS = [
  columns.episodeId,
  columns.positionSec,
  columns.firstPlayedAt,
  columns.firstCompletedAt,
  columns.lastPlayedAt,
].join(", ");

/**
 * 時刻 merge の共通部。first_played_at は先勝ち、position_sec と last_played_at は同じ側のペアで後勝ち。
 *
 * why: `DO UPDATE SET` の右辺で修飾なしの列は更新前の行を指す（`excluded.` が今回の入力）。
 * このため position_sec と last_played_at を同じ文で更新しても比較は更新前の last_played_at で行われ、
 * ペアが割れない。同時刻（`>=`）は再送・後着を採る。
 */
const MERGE_ASSIGNMENTS = [
  `${columns.firstPlayedAt} = MIN(${columns.firstPlayedAt}, excluded.${columns.firstPlayedAt})`,
  `${columns.positionSec} = CASE WHEN excluded.${columns.lastPlayedAt} >= ${columns.lastPlayedAt} THEN excluded.${columns.positionSec} ELSE ${columns.positionSec} END`,
  `${columns.lastPlayedAt} = MAX(${columns.lastPlayedAt}, excluded.${columns.lastPlayedAt})`,
];

/** complete だけが first_completed_at を先勝ちで埋める。NULL（未完走）は COALESCE で今回値に譲る。 */
const COMPLETE_ASSIGNMENT = `${columns.firstCompletedAt} = MIN(COALESCE(${columns.firstCompletedAt}, excluded.${columns.firstCompletedAt}), excluded.${columns.firstCompletedAt})`;

/**
 * 勝ち側の first* を RETURNING で受ける 1 文 upsert。bind は (?1 episodeId, ?2 positionSec, ?3 clientAt)。
 * 行なしなら first_played_at = last_played_at = clientAt で作る。
 */
function upsertReturningFirstSql(
  firstCompletedAtOnInsert: "NULL" | "?3",
  assignments: readonly string[],
): string {
  return [
    `INSERT INTO ${TABLE} (${ALL_COLUMNS})`,
    `VALUES (?1, ?2, ?3, ${firstCompletedAtOnInsert}, ?3)`,
    `ON CONFLICT(${columns.episodeId}) DO UPDATE SET ${assignments.join(", ")}`,
    `RETURNING ${columns.firstPlayedAt}, ${columns.firstCompletedAt}`,
  ].join(" ");
}

const WRITE_SQL = upsertReturningFirstSql("NULL", MERGE_ASSIGNMENTS);
const COMPLETE_SQL = upsertReturningFirstSql("?3", [...MERGE_ASSIGNMENTS, COMPLETE_ASSIGNMENT]);

/**
 * since より後（厳密に大きい）の行を、更新の古い順・同時刻は episodeId 順で返す。
 *
 * why: 時刻列は UTC 固定幅へ正規化済みのため文字列比較が時系列と一致する。並びを決定的にし、
 * caller が末尾の lastPlayedAt を次回の since に使えるよう昇順にする。
 */
const LIST_UPDATED_SINCE_SQL = `SELECT ${ALL_COLUMNS} FROM ${TABLE} WHERE ${columns.lastPlayedAt} > ?1 ORDER BY ${columns.lastPlayedAt} ASC, ${columns.episodeId} ASC`;

function selectByEpisodeIdsSql(count: number): string {
  const placeholders = Array.from({ length: count }, () => "?").join(", ");
  return `SELECT ${ALL_COLUMNS} FROM ${TABLE} WHERE ${columns.episodeId} IN (${placeholders})`;
}

function chunksOf<T>(items: readonly T[], size: number): T[][] {
  const chunks: T[][] = [];
  for (let start = 0; start < items.length; start += size) {
    chunks.push(items.slice(start, start + size));
  }
  return chunks;
}

function stringColumn(row: D1Row, column: string): string {
  const value = row[column];
  if (typeof value !== "string") {
    throw new D1Error(`D1 の進捗行の列 ${column} が文字列でない`);
  }
  return value;
}

function nullableStringColumn(row: D1Row, column: string): string | null {
  return row[column] === null ? null : stringColumn(row, column);
}

function numberColumn(row: D1Row, column: string): number {
  const value = row[column];
  if (typeof value !== "number") {
    throw new D1Error(`D1 の進捗行の列 ${column} が数値でない`);
  }
  return value;
}

function toUtcIsoString(timestamp: string): string {
  return new Date(timestamp).toISOString();
}

function toEpisodeProgress(row: D1Row): EpisodeProgress {
  return {
    positionSec: numberColumn(row, columns.positionSec),
    firstPlayedAt: stringColumn(row, columns.firstPlayedAt),
    firstCompletedAt: nullableStringColumn(row, columns.firstCompletedAt),
    lastPlayedAt: stringColumn(row, columns.lastPlayedAt),
  };
}

/**
 * D1 binding（`D1DatabaseBinding`）で `ProgressRepository` を満たす本番 Driven Adapter。
 *
 * 書込は勝ち側を RETURNING で受ける 1 文 upsert に閉じ、read-then-write の競合も補償 rollback も作らない。
 *
 * why: D1 の TEXT 列は `MIN` も辞書順比較も offset 混在（`+09:00` と `Z`）で時系列とずれる。契約は `Z` /
 * `±hh:mm` の両方を許すため、受理時に UTC 固定幅の ISO へ正規化して保存し、応答の時刻も UTC `Z` 表記にする。
 *
 * @require deps.database は migration 適用済みの進捗表を持つ D1 binding
 * @require clientAt / since は契約 schema を通った ISO-8601。不正値は D1 を呼ぶ前に RangeError で止まる
 * @ensure D1 の例外・RETURNING の欠落・列型の不一致は D1Error を throw する。adapter 内で再試行しない
 * @invariant SQL 全文と bind 値を Error message に含めない
 */
export class D1ProgressRepository implements ProgressRepository {
  private readonly database: D1DatabaseBinding;

  constructor(deps: { database: D1DatabaseBinding }) {
    this.database = deps.database;
  }

  async getByEpisodeIds(
    episodeIds: readonly string[],
  ): Promise<ReadonlyMap<string, EpisodeProgress>> {
    const progressByEpisodeId = new Map<string, EpisodeProgress>();
    for (const chunk of chunksOf(episodeIds, D1_MAX_BOUND_PARAMETERS_PER_QUERY)) {
      const rows = await this.queryAll(selectByEpisodeIdsSql(chunk.length), chunk);
      for (const row of rows) {
        progressByEpisodeId.set(stringColumn(row, columns.episodeId), toEpisodeProgress(row));
      }
    }
    return progressByEpisodeId;
  }

  async writeProgress(command: ProgressWriteCommand): Promise<ProgressWriteResult> {
    return this.upsert(WRITE_SQL, command);
  }

  async completeProgress(command: ProgressWriteCommand): Promise<ProgressWriteResult> {
    return this.upsert(COMPLETE_SQL, command);
  }

  async listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]> {
    const rows = await this.queryAll(LIST_UPDATED_SINCE_SQL, [toUtcIsoString(since)]);
    return rows.map((row) => ({
      episodeId: stringColumn(row, columns.episodeId),
      progress: toEpisodeProgress(row),
    }));
  }

  private async upsert(sql: string, command: ProgressWriteCommand): Promise<ProgressWriteResult> {
    const clientAt = toUtcIsoString(command.clientAt);
    let row: D1Row | null;
    try {
      row = await this.database
        .prepare(sql)
        .bind(command.episodeId, command.positionSec, clientAt)
        .first();
    } catch (cause) {
      throw new D1Error("D1 の進捗書込に失敗", { cause });
    }
    if (row === null) {
      throw new D1Error("D1 の進捗書込が勝ち側の行を返さなかった");
    }
    return {
      firstPlayedAt: stringColumn(row, columns.firstPlayedAt),
      firstCompletedAt: nullableStringColumn(row, columns.firstCompletedAt),
    };
  }

  private async queryAll(sql: string, values: readonly unknown[]): Promise<D1Row[]> {
    try {
      const { results } = await this.database
        .prepare(sql)
        .bind(...values)
        .all();
      return results;
    } catch (cause) {
      throw new D1Error("D1 の進捗読取に失敗", { cause });
    }
  }
}
