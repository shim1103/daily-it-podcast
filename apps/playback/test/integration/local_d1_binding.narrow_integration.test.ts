/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers D1 binding
 * Double: 本番 remote D1 は使わない（remoteBindings: false）
 *
 * @require createLocalD1Binding が実 proxy を起動し、migration を適用した空の D1 を得られる
 * @ensure prepare→bind→run→first で binding が観測可能（D1 adapter が刺せる入口）
 */
import { describe, expect, it } from "vitest";
import {
  EPISODE_PROGRESS_TABLE,
  episodeProgressColumns,
} from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { applyEpisodeProgressMigration } from "../support/apply-episode-progress-migration.ts";
import { createLocalD1Binding } from "../support/create-local-d1-binding.ts";

const columns = episodeProgressColumns;

const INSERT_PROGRESS = `INSERT INTO ${EPISODE_PROGRESS_TABLE} (${columns.episodeId}, ${columns.positionSec}, ${columns.firstPlayedAt}, ${columns.firstCompletedAt}, ${columns.lastPlayedAt}) VALUES (?, ?, ?, ?, ?)`;
const SELECT_PROGRESS = `SELECT * FROM ${EPISODE_PROGRESS_TABLE} WHERE ${columns.episodeId} = ?`;

describe("createLocalD1Binding", () => {
  it("reports_one_change_on_run_when_inserting_row_with_bound_values", async () => {
    // Given: getPlatformProxy local binding infra と migration 適用済みの空の表
    const handle = await createLocalD1Binding();

    try {
      await applyEpisodeProgressMigration(handle.database);

      // When: bind した値で 1 行を INSERT する
      const written = await handle.database
        .prepare(INSERT_PROGRESS)
        .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
        .run();

      // Then: 実 binding として書込が成功し 1 行変わる
      expect(written.success).toBe(true);
      expect(written.meta.changes).toBe(1);
    } finally {
      await handle.dispose();
    }
  });

  it("returns_inserted_row_on_first_when_selecting_by_bound_episode_id", async () => {
    // Given: 1 行を INSERT 済みの表
    const handle = await createLocalD1Binding();

    try {
      await applyEpisodeProgressMigration(handle.database);
      await handle.database
        .prepare(INSERT_PROGRESS)
        .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
        .run();

      // When: bind した episode_id で SELECT して first を取る
      const row = await handle.database.prepare(SELECT_PROGRESS).bind("ep-1").first();

      // Then: INSERT した値が列名どおりに読める（NULL も NULL のまま）
      expect(row).toEqual({
        [columns.episodeId]: "ep-1",
        [columns.positionSec]: 12.5,
        [columns.firstPlayedAt]: "2026-10-01T00:00:00Z",
        [columns.firstCompletedAt]: null,
        [columns.lastPlayedAt]: "2026-10-01T00:00:00Z",
      });
    } finally {
      await handle.dispose();
    }
  });

  it("returns_null_on_first_when_no_row_matches_bound_episode_id", async () => {
    // Given: 1 行を INSERT 済みの表
    const handle = await createLocalD1Binding();

    try {
      await applyEpisodeProgressMigration(handle.database);
      await handle.database
        .prepare(INSERT_PROGRESS)
        .bind("ep-1", 12.5, "2026-10-01T00:00:00Z", null, "2026-10-01T00:00:00Z")
        .run();

      // When: 存在しない episode_id で SELECT して first を取る
      const row = await handle.database.prepare(SELECT_PROGRESS).bind("ep-unknown").first();

      // Then: 一致行が無いので null
      expect(row).toBeNull();
    } finally {
      await handle.dispose();
    }
  });
});
