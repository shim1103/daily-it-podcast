// @vitest-environment node
/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers D1 binding（実 SQLite）
 * Double: 本番 remote D1 は使わない（remoteBindings: false）
 *
 * 実 SQL の全列置換（新規 INSERT と既存行の置換）・pull の境界と並び・実 D1 の bind 数上限・
 * 表なし失敗を所有する。勝敗の merge 規則は Application（merge-progress の sociable unit）が所有する。
 * binding 呼び出しの形・Error 写像・再試行なしは sociable unit が所有する。
 *
 * @require createLocalD1Binding が実 proxy を起動し、applyEpisodeProgressMigration で表を作れる
 * @ensure D1ProgressRepository が実 SQLite に対して Port の永続契約（渡された行をそのまま保存し、読み返せる）を満たす
 */
import { afterAll, beforeAll, beforeEach, describe, expect, it } from "vitest";
import type { ProgressUpsertRow } from "../../worker/src/application/ports/progress-repository.ts";
import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";
import { D1Error } from "../../worker/src/infrastructure/d1/d1-error.ts";
import { D1ProgressRepository } from "../../worker/src/infrastructure/d1/d1-progress-repository.ts";
import {
  D1_MAX_BOUND_PARAMETERS_PER_QUERY,
  EPISODE_PROGRESS_TABLE,
} from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { applyEpisodeProgressMigration } from "../support/apply-episode-progress-migration.ts";
import {
  type LocalD1BindingHandle,
  createLocalD1Binding,
} from "../support/create-local-d1-binding.ts";

// why: 起動に失敗すると handle は未代入のまま afterAll に来る。dispose の TypeError で元の失敗原因を隠さない
let handle: LocalD1BindingHandle | undefined;
let database: D1DatabaseBinding;
let repository: D1ProgressRepository;

beforeAll(async () => {
  handle = await createLocalD1Binding();
  database = handle.database;
  await applyEpisodeProgressMigration(database);
  repository = new D1ProgressRepository({ database });
});

afterAll(async () => {
  await handle?.dispose();
});

beforeEach(async () => {
  await database.prepare(`DELETE FROM ${EPISODE_PROGRESS_TABLE}`).run();
});

function progressRow(
  episodeId: string,
  overrides: Partial<ProgressUpsertRow> = {},
): ProgressUpsertRow {
  return {
    episodeId,
    positionSec: 10,
    firstPlayedAt: "2026-10-01T00:00:10.000Z",
    firstCompletedAt: null,
    lastPlayedAt: "2026-10-01T00:00:10.000Z",
    ...overrides,
  };
}

async function progressOf(episodeId: string) {
  return (await repository.getByEpisodeIds([episodeId])).get(episodeId);
}

describe("D1ProgressRepository on local D1", () => {
  describe("upsertProgress", () => {
    it("inserts_row_with_all_columns_as_given_when_no_row_exists", async () => {
      // Given: 行の無い episode と、完走済みの row
      const row = progressRow("ep-1", {
        positionSec: 60.5,
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
        lastPlayedAt: "2026-10-01T00:02:00.000Z",
      });

      // When: upsert する
      await repository.upsertProgress(row);

      // Then: 全列が渡した値のまま読み返せる
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 60.5,
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
        lastPlayedAt: "2026-10-01T00:02:00.000Z",
      });
    });

    it("replaces_every_column_as_given_without_merging_when_row_exists", async () => {
      // Given: 完走済みで最終再生が新しい行
      await repository.upsertProgress(
        progressRow("ep-1", {
          positionSec: 60,
          firstPlayedAt: "2026-10-01T00:00:10.000Z",
          firstCompletedAt: "2026-10-01T00:01:00.000Z",
          lastPlayedAt: "2026-10-01T00:02:00.000Z",
        }),
      );

      // When: 全列で既存より古い・未完走の row を upsert する
      await repository.upsertProgress(
        progressRow("ep-1", {
          positionSec: 5,
          firstPlayedAt: "2026-10-01T00:00:20.000Z",
          firstCompletedAt: null,
          lastPlayedAt: "2026-10-01T00:00:30.000Z",
        }),
      );

      // Then: 勝敗を決めず、渡した値で全列が置き換わる（firstCompletedAt も null に戻る）
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 5,
        firstPlayedAt: "2026-10-01T00:00:20.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:30.000Z",
      });
    });

    it("leaves_other_episode_rows_untouched_when_one_row_is_replaced", async () => {
      // Given: 2 episode の行
      await repository.upsertProgress(progressRow("ep-1", { positionSec: 1 }));
      await repository.upsertProgress(progressRow("ep-2", { positionSec: 2 }));

      // When: ep-1 だけ置き換える
      await repository.upsertProgress(progressRow("ep-1", { positionSec: 99 }));

      // Then: ep-2 は変わらない
      expect((await progressOf("ep-1"))?.positionSec).toBe(99);
      expect((await progressOf("ep-2"))?.positionSec).toBe(2);
    });
  });

  describe("getByEpisodeIds", () => {
    it("returns_only_existing_rows_when_some_episode_ids_have_no_row", async () => {
      // Given: ep-1 だけ行がある
      await repository.upsertProgress(progressRow("ep-1"));

      // When: 行のある ID と無い ID を一緒に取得する
      const got = await repository.getByEpisodeIds(["ep-1", "ep-unknown"]);

      // Then: 行のある episode だけが載る
      expect([...got.keys()]).toEqual(["ep-1"]);
    });

    it("finds_rows_across_chunks_when_episode_ids_exceed_real_bind_limit", async () => {
      // Given: 実 D1 の bind 数上限を超える数の ID と、チャンクをまたぐ 3 行
      const episodeIds = Array.from(
        { length: D1_MAX_BOUND_PARAMETERS_PER_QUERY * 2 + 50 },
        (_, index) => `ep-${index}`,
      );
      const seeded = ["ep-5", `ep-${D1_MAX_BOUND_PARAMETERS_PER_QUERY + 20}`, "ep-249"];
      for (const episodeId of seeded) {
        await repository.upsertProgress(progressRow(episodeId));
      }

      // When: 全 ID で取得する
      const got = await repository.getByEpisodeIds(episodeIds);

      // Then: 上限エラーにならず、全チャンクの行が揃う
      expect([...got.keys()].sort()).toEqual([...seeded].sort());
    });
  });

  describe("listUpdatedSince", () => {
    it("excludes_row_whose_last_played_at_equals_since", async () => {
      // Given: last_played_at が 00:00:10 / 00:00:20 の 2 行
      await repository.upsertProgress(
        progressRow("ep-1", { lastPlayedAt: "2026-10-01T00:00:10.000Z" }),
      );
      await repository.upsertProgress(
        progressRow("ep-2", { lastPlayedAt: "2026-10-01T00:00:20.000Z" }),
      );

      // When: since = 00:00:10 で pull する
      const got = await repository.listUpdatedSince("2026-10-01T00:00:10.000Z");

      // Then: 厳密に後の行だけが返る
      expect(got.map((entry) => entry.episodeId)).toEqual(["ep-2"]);
    });

    it("orders_by_last_played_at_then_episode_id", async () => {
      // Given: 書込順と更新時刻順が食い違い、同時刻の行を含む 3 行
      await repository.upsertProgress(
        progressRow("ep-b", { lastPlayedAt: "2026-10-01T00:00:20.000Z" }),
      );
      await repository.upsertProgress(
        progressRow("ep-c", { lastPlayedAt: "2026-10-01T00:00:10.000Z" }),
      );
      await repository.upsertProgress(
        progressRow("ep-a", { lastPlayedAt: "2026-10-01T00:00:20.000Z" }),
      );

      // When: pull する
      const got = await repository.listUpdatedSince("2026-10-01T00:00:00.000Z");

      // Then: 古い順、同時刻は episodeId 昇順
      expect(got.map((entry) => entry.episodeId)).toEqual(["ep-c", "ep-a", "ep-b"]);
    });

    it("returns_empty_array_when_no_row_is_newer_than_since", async () => {
      // Given: 1 行
      await repository.upsertProgress(progressRow("ep-1"));

      // When: 全行より後の since で pull する
      const got = await repository.listUpdatedSince("2026-10-02T00:00:00.000Z");

      // Then: 空配列
      expect(got).toEqual([]);
    });
  });

  describe("failure", () => {
    it("throws_d1_error_with_cause_when_progress_table_does_not_exist", async () => {
      // Given: migration 未適用の空の D1（実 D1 が失敗を返す）
      const bare = await createLocalD1Binding();
      try {
        const bareRepository = new D1ProgressRepository({ database: bare.database });

        // When: 書込する
        const got = await bareRepository
          .upsertProgress(progressRow("ep-1"))
          .catch((error: unknown) => error);

        // Then: D1Error に畳まれ、実 D1 の失敗が cause に残る
        expect(got).toBeInstanceOf(D1Error);
        expect((got as D1Error).cause).toBeInstanceOf(Error);
      } finally {
        await bare.dispose();
      }
    });
  });
});
