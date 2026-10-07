import type { EpisodeProgress } from "../../entities/models/episode-progress.ts";
import type {
  GetByEpisodeIdsArgs,
  GetByEpisodeIdsResult,
  GetProgressWithVersionArgs,
  GetProgressWithVersionResult,
  InsertProgressIfAbsentArgs,
  ListChangedAfterArgs,
  ListChangedAfterResult,
  ProgressConditionalWriteResult,
  ProgressRepository,
  ProgressRow,
  ProgressUpdatedEntry,
  ProgressVersion,
  ReplaceProgressIfVersionArgs,
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

type D1RunResult = Awaited<ReturnType<D1PreparedStatementBinding["run"]>>;

const columns = episodeProgressColumns;
const TABLE = EPISODE_PROGRESS_TABLE;

const NON_KEY_COLUMNS = [
  columns.positionSec,
  columns.firstPlayedAt,
  columns.firstCompletedAt,
  columns.lastPlayedAt,
] as const;

const ALL_COLUMNS = [columns.episodeId, ...NON_KEY_COLUMNS].join(", ");

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
  NON_KEY_COLUMNS.map((column) => `${column} = excluded.${column}`).join(", "),
].join(" ");

/** since より後（厳密に大きい）の行を、更新の古い順・同時刻は episodeId 順で返す。 */
const LIST_UPDATED_SINCE_SQL = `SELECT ${ALL_COLUMNS} FROM ${TABLE} WHERE ${columns.lastPlayedAt} > ?1 ORDER BY ${columns.lastPlayedAt} ASC, ${columns.episodeId} ASC`;

/** 現在の最大 `seq` + 1（空表は 1）。索引 `episode_progress_seq_idx` が最大値の取得を効かせる。 */
const NEXT_SEQ_SQL = `(SELECT COALESCE(MAX(${columns.seq}), 0) + 1 FROM ${TABLE})`;

/**
 * 行なしを期待する条件付き挿入。既に行があれば何も書かない（changes 0）。
 * bind は ?1 episodeId, ?2 positionSec, ?3 firstPlayedAt, ?4 firstCompletedAt, ?5 lastPlayedAt。
 */
const INSERT_IF_ABSENT_SQL = [
  `INSERT INTO ${TABLE} (${ALL_COLUMNS}, ${columns.seq})`,
  `VALUES (?1, ?2, ?3, ?4, ?5, ${NEXT_SEQ_SQL})`,
  `ON CONFLICT(${columns.episodeId}) DO NOTHING`,
].join(" ");

/**
 * 期待した版の行だけを置き換える条件付き更新。版が食い違う、または行が無ければ何も書かない（changes 0）。
 * bind は INSERT_IF_ABSENT_SQL の ?1〜?5 に、?6 期待する版を加えたもの。
 */
const UPDATE_IF_VERSION_SQL = [
  `UPDATE ${TABLE} SET`,
  [
    ...NON_KEY_COLUMNS.map((column, index) => `${column} = ?${index + 2}`),
    `${columns.seq} = ${NEXT_SEQ_SQL}`,
  ].join(", "),
  `WHERE ${columns.episodeId} = ?1 AND ${columns.seq} = ?6`,
].join(" ");

const SELECT_WITH_VERSION_COLUMNS = `${ALL_COLUMNS}, ${columns.seq}`;

const SELECT_WITH_VERSION_SQL = `SELECT ${SELECT_WITH_VERSION_COLUMNS} FROM ${TABLE} WHERE ${columns.episodeId} = ?1`;

/** cursor より後（厳密に大きい）の行を、書込順（`seq` 昇順）で返す。 */
const LIST_CHANGED_AFTER_SQL = `SELECT ${SELECT_WITH_VERSION_COLUMNS} FROM ${TABLE} WHERE ${columns.seq} > ?1 ORDER BY ${columns.seq} ASC`;

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

/** ALL_COLUMNS の列順（?1 episodeId, ?2 positionSec, ?3 firstPlayedAt, ?4 firstCompletedAt, ?5 lastPlayedAt）の bind 値。 */
function rowValues(row: ProgressRow): unknown[] {
  return [
    row.episodeId,
    row.positionSec,
    row.firstPlayedAt,
    row.firstCompletedAt,
    row.lastPlayedAt,
  ];
}

function toUpdatedEntry(row: D1Row): ProgressUpdatedEntry {
  return {
    episodeId: stringColumn(row, columns.episodeId),
    progress: toEpisodeProgress(row),
  };
}

function toConditionalWriteResult(result: D1RunResult): ProgressConditionalWriteResult {
  return result.meta.changes === 1 ? "written" : "conflict";
}

function toProgressVersion(row: D1Row): ProgressVersion {
  // why: ProgressVersion は Port が公開する不透明な型で、行の採番列から作れるのはこの adapter だけ
  return numberColumn(row, columns.seq) as ProgressVersion;
}

/**
 * D1 binding（`D1DatabaseBinding`）で `ProgressRepository` を満たす本番 Driven Adapter。
 * 渡された行を保存し、読み返すだけで merge しない。
 *
 * @require deps.database は migration 適用済みの進捗表を持つ D1 binding
 * @require 旧面 `listUpdatedSince` の `since` と行の時刻は、辞書順が時系列と一致する表記（UTC 固定幅の ISO-8601）。契約 schema が入口で保証し、adapter は変換しない（`>` 比較と並びを文字列で行うため）
 * @ensure D1 の例外・`run()` の `success === false`・列型の不一致は D1Error を throw する。adapter 内で再試行しない
 * @invariant `seq` と cursor を数値のまま扱い、文字列へ変換しない
 * @invariant SQL 全文と bind 値を Error message に含めない
 */
export class D1ProgressRepository implements ProgressRepository {
  private readonly database: D1DatabaseBinding;

  constructor(deps: { database: D1DatabaseBinding }) {
    this.database = deps.database;
  }

  async getByEpisodeIds(args: GetByEpisodeIdsArgs): Promise<GetByEpisodeIdsResult> {
    const progressByEpisodeId = new Map<string, EpisodeProgress>();
    for (const chunk of chunksOf(args.episodeIds, D1_MAX_BOUND_PARAMETERS_PER_QUERY)) {
      const rows = await this.queryAll(selectByEpisodeIdsSql(chunk.length), chunk);
      for (const row of rows) {
        progressByEpisodeId.set(stringColumn(row, columns.episodeId), toEpisodeProgress(row));
      }
    }
    return progressByEpisodeId;
  }

  async upsertProgress(row: ProgressRow): Promise<void> {
    await this.runWrite(UPSERT_SQL, rowValues(row));
  }

  async listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]> {
    const rows = await this.queryAll(LIST_UPDATED_SINCE_SQL, [since]);
    return rows.map(toUpdatedEntry);
  }

  async getProgressWithVersion(
    args: GetProgressWithVersionArgs,
  ): Promise<GetProgressWithVersionResult> {
    const row = await this.queryFirst(SELECT_WITH_VERSION_SQL, [args.episodeId]);
    if (row === null) {
      return null;
    }
    return { progress: toEpisodeProgress(row), version: toProgressVersion(row) };
  }

  async insertProgressIfAbsent(
    args: InsertProgressIfAbsentArgs,
  ): Promise<ProgressConditionalWriteResult> {
    const result = await this.runWrite(INSERT_IF_ABSENT_SQL, rowValues(args.row));
    return toConditionalWriteResult(result);
  }

  async replaceProgressIfVersion(
    args: ReplaceProgressIfVersionArgs,
  ): Promise<ProgressConditionalWriteResult> {
    const result = await this.runWrite(UPDATE_IF_VERSION_SQL, [
      ...rowValues(args.row),
      args.expectedVersion,
    ]);
    return toConditionalWriteResult(result);
  }

  async listChangedAfter(args: ListChangedAfterArgs): Promise<ListChangedAfterResult> {
    const rows = await this.queryAll(LIST_CHANGED_AFTER_SQL, [args.cursor]);
    const lastRow = rows.at(-1);
    return {
      cursor: lastRow === undefined ? args.cursor : numberColumn(lastRow, columns.seq),
      entries: rows.map(toUpdatedEntry),
    };
  }

  private async runWrite(sql: string, values: readonly unknown[]): Promise<D1RunResult> {
    let result: D1RunResult;
    try {
      result = await this.database
        .prepare(sql)
        .bind(...values)
        .run();
    } catch (cause) {
      throw new D1Error("D1 の進捗書込に失敗", { cause });
    }
    if (!result.success) {
      throw new D1Error("D1 の進捗書込が成功を返さなかった");
    }
    return result;
  }

  private async queryFirst(sql: string, values: readonly unknown[]): Promise<D1Row | null> {
    try {
      return await this.database
        .prepare(sql)
        .bind(...values)
        .first();
    } catch (cause) {
      throw new D1Error("D1 の進捗読取に失敗", { cause });
    }
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
