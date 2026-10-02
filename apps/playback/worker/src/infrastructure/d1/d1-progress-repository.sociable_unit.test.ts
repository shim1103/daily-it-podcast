import { describe, expect, it } from "vitest";
import { createFakeLocalD1Binding } from "../../../../test/support/create-fake-local-d1-binding.ts";
import type { D1Row } from "./d1-database-binding.ts";
import { D1Error } from "./d1-error.ts";
import { D1ProgressRepository } from "./d1-progress-repository.ts";
import {
  D1_MAX_BOUND_PARAMETERS_PER_QUERY,
  PROGRESS_D1_ADAPTER_MAX_ATTEMPTS,
  episodeProgressColumns,
} from "./progress-d1-constants.ts";

/**
 * scope: Sociable Unit
 * real: D1ProgressRepository
 * double: D1DatabaseBinding（共有 Fake local binding。応答を差し込み、呼び出しを記録する）
 *
 * 勝敗の merge 規則は Application（`merge-progress.sociable_unit.test.ts`）が所有する。
 * 全列置換・pull の境界と並び・bind 数上限の実 SQLite での意味は
 * `test/integration/d1_progress_repository.narrow_integration.test.ts` が所有する。
 */

const columns = episodeProgressColumns;

function progressRow(overrides: D1Row = {}): D1Row {
  return {
    [columns.episodeId]: "ep-1",
    [columns.positionSec]: 12.5,
    [columns.firstPlayedAt]: "2026-10-01T00:00:00.000Z",
    [columns.firstCompletedAt]: null,
    [columns.lastPlayedAt]: "2026-10-01T00:05:00.000Z",
    ...overrides,
  };
}

describe("D1ProgressRepository", () => {
  describe("getByEpisodeIds", () => {
    it("returns_progress_keyed_by_episode_id_when_rows_exist", async () => {
      // Given: ep-1 だけ行がある D1
      const { database, calls } = await createFakeLocalD1Binding({
        all: async () => [progressRow({ [columns.firstCompletedAt]: "2026-10-01T00:04:00.000Z" })],
      });
      const repository = new D1ProgressRepository({ database });

      // When: 行のある episode と無い episode を同時に取得する
      const got = await repository.getByEpisodeIds(["ep-1", "ep-2"]);

      // Then: 行のある episode だけが Map に載り、列が契約型へ写る。bind は episodeId をそのまま渡す
      expect([...got.keys()]).toEqual(["ep-1"]);
      expect(got.get("ep-1")).toEqual({
        positionSec: 12.5,
        firstPlayedAt: "2026-10-01T00:00:00.000Z",
        firstCompletedAt: "2026-10-01T00:04:00.000Z",
        lastPlayedAt: "2026-10-01T00:05:00.000Z",
      });
      expect(calls).toHaveLength(1);
      expect(calls[0]?.values).toEqual(["ep-1", "ep-2"]);
    });

    it("returns_empty_map_when_no_row_exists", async () => {
      // Given: 行が 1 つも無い D1（Fake の既定応答）
      const { database } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 取得する
      const got = await repository.getByEpisodeIds(["ep-1"]);

      // Then: 空 Map（null でない）
      expect(got.size).toBe(0);
    });

    it("returns_empty_map_without_calling_d1_when_episode_ids_is_empty", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 空配列で取得する
      const got = await repository.getByEpisodeIds([]);

      // Then: D1 を 1 度も呼ばず空 Map を返す
      expect(got.size).toBe(0);
      expect(calls).toHaveLength(0);
    });

    it("splits_query_within_bind_limit_and_merges_results_when_ids_exceed_limit", async () => {
      // Given: 上限を 1 つ超える episodeId と、各 query が bind された先頭の 1 件を返す D1
      const episodeIds = Array.from(
        { length: D1_MAX_BOUND_PARAMETERS_PER_QUERY + 1 },
        (_, index) => `ep-${index}`,
      );
      const { database, calls } = await createFakeLocalD1Binding({
        all: async (call) => [progressRow({ [columns.episodeId]: call.values[0] })],
      });
      const repository = new D1ProgressRepository({ database });

      // When: 取得する
      const got = await repository.getByEpisodeIds(episodeIds);

      // Then: 上限以内の bind で 2 query に分かれ、結果は 1 つの Map に合流する
      expect(calls.map((call) => call.values.length)).toEqual([
        D1_MAX_BOUND_PARAMETERS_PER_QUERY,
        1,
      ]);
      expect([...got.keys()]).toEqual(["ep-0", `ep-${D1_MAX_BOUND_PARAMETERS_PER_QUERY}`]);
    });

    it("sends_single_query_when_ids_equal_bind_limit", async () => {
      // Given: ちょうど上限個数の episodeId と、呼び出しを記録する D1
      const episodeIds = Array.from(
        { length: D1_MAX_BOUND_PARAMETERS_PER_QUERY },
        (_, index) => `ep-${index}`,
      );
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 取得する
      await repository.getByEpisodeIds(episodeIds);

      // Then: 上限ちょうどは分割せず 1 query に収まる（末尾に空の chunk を作らない）
      expect(calls.map((call) => call.values.length)).toEqual([D1_MAX_BOUND_PARAMETERS_PER_QUERY]);
    });

    it("splits_into_full_queries_without_empty_chunk_when_ids_are_twice_bind_limit", async () => {
      // Given: 上限のちょうど 2 倍の episodeId と、呼び出しを記録する D1
      const episodeIds = Array.from(
        { length: D1_MAX_BOUND_PARAMETERS_PER_QUERY * 2 },
        (_, index) => `ep-${index}`,
      );
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 取得する
      await repository.getByEpisodeIds(episodeIds);

      // Then: 上限いっぱいの 2 query だけに分かれ、bind 0 個の query（IN ()）を送らない
      expect(calls.map((call) => call.values.length)).toEqual([
        D1_MAX_BOUND_PARAMETERS_PER_QUERY,
        D1_MAX_BOUND_PARAMETERS_PER_QUERY,
      ]);
    });

    it("throws_d1_error_with_cause_without_retry_when_d1_fails", async () => {
      // Given: 読取が失敗する D1
      const cause = new Error("D1_ERROR: network");
      const { database, calls } = await createFakeLocalD1Binding({
        all: async () => {
          throw cause;
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: 取得する
      const got = repository.getByEpisodeIds(["ep-1"]);

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it("throws_d1_error_without_leaking_values_when_d1_fails", async () => {
      // Given: bind 値を含む message で失敗する D1
      const { database } = await createFakeLocalD1Binding({
        all: async () => {
          throw new Error("failed for ep-secret");
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: 取得する
      const got = await repository.getByEpisodeIds(["ep-secret"]).catch((error: unknown) => error);

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("ep-secret");
      expect((got as D1Error).message).not.toContain("SELECT");
    });
  });

  describe("upsertProgress", () => {
    const row = {
      episodeId: "ep-1",
      positionSec: 42.5,
      firstPlayedAt: "2026-10-01T00:00:00.000Z",
      firstCompletedAt: null,
      lastPlayedAt: "2026-10-01T00:05:00.000Z",
    };

    it("binds_row_columns_in_table_order_with_one_statement", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 未完走の row を upsert する
      await repository.upsertProgress(row);

      // Then: D1 へは 1 文だけ送り（read-then-write にしない）、row の列を値のまま表の列順に bind する
      expect(calls).toHaveLength(1);
      expect(calls[0]?.values).toEqual([
        "ep-1",
        42.5,
        "2026-10-01T00:00:00.000Z",
        null,
        "2026-10-01T00:05:00.000Z",
      ]);
    });

    it("binds_first_completed_at_as_is_when_row_is_completed", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 完走済みの row を upsert する
      await repository.upsertProgress({ ...row, firstCompletedAt: "2026-10-01T00:04:00.000Z" });

      // Then: firstCompletedAt を値のまま bind する
      expect(calls[0]?.values[3]).toBe("2026-10-01T00:04:00.000Z");
    });

    it("replaces_every_column_without_merging_in_sql", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: upsert する
      await repository.upsertProgress(row);

      // Then: 鍵以外の全列を excluded の値へ置き換え、勝敗を決める SQL 関数も返却句も持たない
      const sql = calls[0]?.sql ?? "";
      expect(sql).toContain(`ON CONFLICT(${columns.episodeId}) DO UPDATE SET`);
      for (const column of [
        columns.positionSec,
        columns.firstPlayedAt,
        columns.firstCompletedAt,
        columns.lastPlayedAt,
      ]) {
        expect(sql).toContain(`${column} = excluded.${column}`);
      }
      expect(sql).not.toMatch(/\b(MIN|MAX|CASE|COALESCE|RETURNING)\b/);
    });

    it("resolves_with_no_value_when_d1_run_succeeds", async () => {
      // Given: 書込が成功する D1（Fake の既定応答）
      const { database } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: upsert する
      const got = repository.upsertProgress(row);

      // Then: 応答を返さず完了する
      await expect(got).resolves.toBeUndefined();
    });

    it("throws_d1_error_with_cause_without_retry_when_d1_throws", async () => {
      // Given: 書込が throw する D1
      const cause = new Error("D1_ERROR: unavailable");
      const { database, calls } = await createFakeLocalD1Binding({
        run: async () => {
          throw cause;
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: upsert する
      const got = repository.upsertProgress(row);

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it("throws_d1_error_without_retry_when_d1_reports_unsuccessful_run", async () => {
      // Given: success が false の結果を返す D1
      const { database, calls } = await createFakeLocalD1Binding({
        run: async () => ({ success: false, meta: { changes: 0 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: upsert する
      const got = repository.upsertProgress(row);

      // Then: 書けたように見せず D1Error を throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it.each([
      [
        "d1_throws",
        {
          run: async () => {
            throw new Error("failed for ep-secret");
          },
        },
      ],
      [
        "d1_reports_unsuccessful_run",
        { run: async () => ({ success: false, meta: { changes: 0 } }) },
      ],
    ])("throws_d1_error_without_leaking_values_when_%s", async (_name, responses) => {
      // Given: 失敗する D1
      const { database } = await createFakeLocalD1Binding(responses);
      const repository = new D1ProgressRepository({ database });

      // When: bind 値を含む row を upsert する
      const got = await repository
        .upsertProgress({ ...row, episodeId: "ep-secret" })
        .catch((error: unknown) => error);

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("ep-secret");
      expect((got as D1Error).message).not.toContain("INSERT");
    });
  });

  describe("listUpdatedSince", () => {
    it("returns_entries_in_row_order_when_rows_exist", async () => {
      // Given: 2 行を返す D1
      const { database } = await createFakeLocalD1Binding({
        all: async () => [
          progressRow({ [columns.episodeId]: "ep-1" }),
          progressRow({
            [columns.episodeId]: "ep-2",
            [columns.positionSec]: 3,
            [columns.firstCompletedAt]: "2026-10-01T00:06:00.000Z",
          }),
        ],
      });
      const repository = new D1ProgressRepository({ database });

      // When: pull する
      const got = await repository.listUpdatedSince("2026-09-30T00:00:00.000Z");

      // Then: 行の並びのまま episodeId と契約型の progress に写る
      expect(got).toEqual([
        {
          episodeId: "ep-1",
          progress: {
            positionSec: 12.5,
            firstPlayedAt: "2026-10-01T00:00:00.000Z",
            firstCompletedAt: null,
            lastPlayedAt: "2026-10-01T00:05:00.000Z",
          },
        },
        {
          episodeId: "ep-2",
          progress: {
            positionSec: 3,
            firstPlayedAt: "2026-10-01T00:00:00.000Z",
            firstCompletedAt: "2026-10-01T00:06:00.000Z",
            lastPlayedAt: "2026-10-01T00:05:00.000Z",
          },
        },
      ]);
    });

    it("returns_empty_array_when_no_row_is_updated", async () => {
      // Given: 行が 1 つも無い D1（Fake の既定応答）
      const { database } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: pull する
      const got = await repository.listUpdatedSince("2026-09-30T00:00:00.000Z");

      // Then: 空配列（null でない）
      expect(got).toEqual([]);
    });

    it("binds_since_as_is_without_normalizing_when_since_has_offset", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: +09:00 の since で pull する
      await repository.listUpdatedSince("2026-10-01T09:00:00+09:00");

      // Then: 時刻を変換せず値のまま bind して 1 query だけ送る
      expect(calls).toHaveLength(1);
      expect(calls[0]?.values).toEqual(["2026-10-01T09:00:00+09:00"]);
    });

    it("throws_d1_error_with_cause_without_retry_when_d1_fails", async () => {
      // Given: 読取が失敗する D1
      const cause = new Error("D1_ERROR: unavailable");
      const { database, calls } = await createFakeLocalD1Binding({
        all: async () => {
          throw cause;
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: pull する
      const got = repository.listUpdatedSince("2026-09-30T00:00:00.000Z");

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it.each([
      ["position_sec is not a number", { [columns.positionSec]: "12" }],
      ["first_played_at is not a string", { [columns.firstPlayedAt]: null }],
      ["first_completed_at is neither null nor a string", { [columns.firstCompletedAt]: 1 }],
      ["episode_id is not a string", { [columns.episodeId]: 1 }],
    ])("throws_d1_error_when_row_is_malformed_because_%s", async (_name, overrides) => {
      // Given: 列型が期待と違う行を返す D1
      const { database } = await createFakeLocalD1Binding({
        all: async () => [progressRow(overrides)],
      });
      const repository = new D1ProgressRepository({ database });

      // When / Then: 契約型として返さず D1Error
      await expect(repository.listUpdatedSince("2026-09-30T00:00:00.000Z")).rejects.toBeInstanceOf(
        D1Error,
      );
    });
  });
});
