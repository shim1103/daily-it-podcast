import type {
  ProgressRepository,
  ProgressUpdatedEntry,
  ProgressUpsertRow,
} from "../../application/ports/progress-repository.ts";
import type {
  D1DatabaseBinding,
  D1PreparedStatementBinding,
  D1Row,
} from "./d1-database-binding.ts";
import { D1Error } from "./d1-error.ts";
import {
  D1_MAX_BOUND_PARAMETERS_PER_QUERY,
  EPISODE_PROGRESS_TABLE,
  episodeProgressColumns,
} from "./progress-d1-constants.ts";

type EpisodeProgress = ProgressUpdatedEntry["progress"];
type D1RunResult = Awaited<ReturnType<D1PreparedStatementBinding["run"]>>;

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
 * 鍵以外の全列を今回の入力（`excluded`）へ置き換える 1 文 upsert。bind は ALL_COLUMNS の列順
 * (?1 episodeId, ?2 positionSec, ?3 firstPlayedAt, ?4 firstCompletedAt, ?5 lastPlayedAt)。
 *
 * why: 先勝ち・後勝ちの merge は Application が持つ（Decision 2026-10-01T18:54:14）。SQL で勝敗を
 * 決めると規則が SQL と TS の 2 箇所に分かれるため、adapter は渡された行をそのまま保存する。
 */
const UPSERT_SQL = [
  `INSERT INTO ${TABLE} (${ALL_COLUMNS})`,
  "VALUES (?1, ?2, ?3, ?4, ?5)",
  `ON CONFLICT(${columns.episodeId}) DO UPDATE SET`,
  [columns.positionSec, columns.firstPlayedAt, columns.firstCompletedAt, columns.lastPlayedAt]
    .map((column) => `${column} = excluded.${column}`)
    .join(", "),
].join(" ");

// todo: 契約境界で時刻を UTC 固定幅へ正規化するまで、offset 混在の入力では `>` 境界と並びが時系列と一致しない。正規化を入れたらこの comment を消す
/** since より後（厳密に大きい）の行を、更新の古い順・同時刻は episodeId 順で返す。 */
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
 * 渡された行を保存し、読み返すだけで merge しない。
 *
 * @require deps.database は migration 適用済みの進捗表を持つ D1 binding
 * @require 時刻は辞書順が時系列と一致する表記（UTC 固定幅の ISO-8601）で渡される。adapter は変換しない（`>` 比較と並びを文字列で行うため）
 * @ensure D1 の例外・`run()` の `success === false`・列型の不一致は D1Error を throw する。adapter 内で再試行しない
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

  async upsertProgress(row: ProgressUpsertRow): Promise<void> {
    let result: D1RunResult;
    try {
      result = await this.database
        .prepare(UPSERT_SQL)
        .bind(
          row.episodeId,
          row.positionSec,
          row.firstPlayedAt,
          row.firstCompletedAt,
          row.lastPlayedAt,
        )
        .run();
    } catch (cause) {
      throw new D1Error("D1 の進捗書込に失敗", { cause });
    }
    if (!result.success) {
      throw new D1Error("D1 の進捗書込が成功を返さなかった");
    }
  }

  async listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]> {
    const rows = await this.queryAll(LIST_UPDATED_SINCE_SQL, [since]);
    return rows.map((row) => ({
      episodeId: stringColumn(row, columns.episodeId),
      progress: toEpisodeProgress(row),
    }));
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
