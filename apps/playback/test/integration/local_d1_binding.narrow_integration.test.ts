// @vitest-environment node
/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers D1 binding
 * Double: 本番 remote D1 は使わない（remoteBindings: false）
 *
 * @require createLocalD1Binding が実 proxy を起動し、migration を適用した空の D1 を得られる
 * @ensure prepare→bind→run／first／all で binding が観測可能（D1 adapter が刺せる入口。adapter が呼ぶ first と all を実 binding で確かめる）
 */
import { afterAll, beforeAll, beforeEach, describe, expect, it } from "vitest";
import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";
import {
  EPISODE_PROGRESS_TABLE,
  episodeProgressColumns,
} from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { applyEpisodeProgressMigration } from "../support/apply-episode-progress-migration.ts";
import {
  type LocalD1BindingHandle,
  createLocalD1Binding,
} from "../support/create-local-d1-binding.ts";

const columns = episodeProgressColumns;

const INSERT_PROGRESS = `INSERT INTO ${EPISODE_PROGRESS_TABLE} (${columns.episodeId}, ${columns.positionSec}, ${columns.firstPlayedAt}, ${columns.firstCompletedAt}, ${columns.lastPlayedAt}) VALUES (?, ?, ?, ?, ?)`;
const SELECT_PROGRESS = `SELECT * FROM ${EPISODE_PROGRESS_TABLE} WHERE ${columns.episodeId} = ?`;
const SELECT_ALL_PROGRESS = `SELECT * FROM ${EPISODE_PROGRESS_TABLE} ORDER BY ${columns.episodeId}`;

// why: proxy の起動は重く、test ごとに起動すると既定 timeout に余裕が薄い。1 回だけ起動し、分離は beforeEach の DELETE で担保する。
//   起動に失敗すると handle は未代入のまま afterAll に来る。dispose の TypeError で元の失敗原因を隠さない
let handle: LocalD1BindingHandle | undefined;
let database: D1DatabaseBinding;

beforeAll(async () => {
  handle = await createLocalD1Binding();
  database = handle.database;
  await applyEpisodeProgressMigration(database);
});

afterAll(async () => {
  await handle?.dispose();
});

beforeEach(async () => {
  await database.prepare(`DELETE FROM ${EPISODE_PROGRESS_TABLE}`).run();
});

describe("createLocalD1Binding", () => {
  it("reports_one_change_on_run_when_inserting_row_with_bound_values", async () => {
    // Given: getPlatformProxy local binding infra と migration 適用済みの空の表

    // When: bind した値で 1 行を INSERT する
    const written = await database
      .prepare(INSERT_PROGRESS)
      .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
      .run();

    // Then: 実 binding として書込が成功し 1 行変わる
    expect(written.success).toBe(true);
    expect(written.meta.changes).toBe(1);
  });

  it("returns_inserted_row_on_first_when_selecting_by_bound_episode_id", async () => {
    // Given: 1 行を INSERT 済みの表
    await database
      .prepare(INSERT_PROGRESS)
      .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
      .run();

    // When: bind した episode_id で SELECT して first を取る
    const row = await database.prepare(SELECT_PROGRESS).bind("ep-1").first();

    // Then: INSERT した値が列名どおりに読める（NULL も NULL のまま）
    expect(row).toEqual({
      [columns.episodeId]: "ep-1",
      [columns.positionSec]: 12.5,
      [columns.firstPlayedAt]: "2026-10-01T00:00:00Z",
      [columns.firstCompletedAt]: null,
      [columns.lastPlayedAt]: "2026-10-01T00:00:00Z",
    });
  });

  it("returns_null_on_first_when_no_row_matches_bound_episode_id", async () => {
    // Given: 1 行を INSERT 済みの表
    await database
      .prepare(INSERT_PROGRESS)
      .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
      .run();

    // When: 存在しない episode_id で SELECT して first を取る
    const row = await database.prepare(SELECT_PROGRESS).bind("ep-unknown").first();

    // Then: 一致行が無いので null
    expect(row).toBeNull();
  });

  it("returns_all_inserted_rows_in_results_on_all_when_selecting_multiple_rows", async () => {
    // Given: 2 行を INSERT 済みの表
    await database
      .prepare(INSERT_PROGRESS)
      .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
      .run();
    await database
      .prepare(INSERT_PROGRESS)
      .bind("ep-2", 30, "2026-10-01T01:00:00Z", "2026-10-01T02:00:00Z", "2026-10-01T02:00:00Z")
      .run();

    // When: 全行を SELECT して all を取る
    const rows = await database.prepare(SELECT_ALL_PROGRESS).all();

    // Then: results に全行が列名どおり episode_id 順で入る
    expect(rows.results).toEqual([
      {
        [columns.episodeId]: "ep-1",
        [columns.positionSec]: 12.5,
        [columns.firstPlayedAt]: "2026-10-01T00:00:00Z",
        [columns.firstCompletedAt]: null,
        [columns.lastPlayedAt]: "2026-10-01T00:00:00Z",
      },
      {
        [columns.episodeId]: "ep-2",
        [columns.positionSec]: 30,
        [columns.firstPlayedAt]: "2026-10-01T01:00:00Z",
        [columns.firstCompletedAt]: "2026-10-01T02:00:00Z",
        [columns.lastPlayedAt]: "2026-10-01T02:00:00Z",
      },
    ]);
  });
});
