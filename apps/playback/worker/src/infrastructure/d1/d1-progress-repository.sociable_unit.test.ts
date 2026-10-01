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
 * merge の SQL 意味（先勝ち / 後勝ち / 冪等 / 境界）は実 SQLite でしか観測できないため、
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

  describe.each(["writeProgress", "completeProgress"] as const)("%s", (operation) => {
    const command = {
      episodeId: "ep-1",
      positionSec: 42.5,
      clientAt: "2026-10-01T09:00:00+09:00",
    };

    it("returns_winning_first_timestamps_from_returning_row_with_one_statement", async () => {
      // Given: 勝ち側の first* を RETURNING で返す D1
      const { database, calls } = await createFakeLocalD1Binding({
        first: async () => ({
          [columns.firstPlayedAt]: "2026-09-30T23:00:00.000Z",
          [columns.firstCompletedAt]: "2026-09-30T23:30:00.000Z",
        }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 書込する
      const got = await repository[operation](command);

      // Then: RETURNING の first* をそのまま返し、D1 へは 1 文だけ送る（read-then-write にしない）
      expect(got).toEqual({
        firstPlayedAt: "2026-09-30T23:00:00.000Z",
        firstCompletedAt: "2026-09-30T23:30:00.000Z",
      });
      expect(calls).toHaveLength(1);
    });

    it("returns_null_first_completed_at_when_returning_row_has_null", async () => {
      // Given: 未完走の行を返す D1
      const { database } = await createFakeLocalD1Binding({
        first: async () => ({
          [columns.firstPlayedAt]: "2026-09-30T23:00:00.000Z",
          [columns.firstCompletedAt]: null,
        }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 書込する
      const got = await repository[operation](command);

      // Then: firstCompletedAt は null のまま返る
      expect(got.firstCompletedAt).toBeNull();
    });

    it("binds_episode_id_position_and_utc_normalized_client_at_when_client_at_has_offset", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding({
        first: async () => ({
          [columns.firstPlayedAt]: "2026-10-01T00:00:00.000Z",
          [columns.firstCompletedAt]: null,
        }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: +09:00 の clientAt で書込する
      await repository[operation](command);

      // Then: 比較可能な UTC 固定幅表現へ揃えて bind する
      expect(calls[0]?.values).toEqual(["ep-1", 42.5, "2026-10-01T00:00:00.000Z"]);
    });

    it("throws_d1_error_with_cause_without_retry_when_d1_fails", async () => {
      // Given: 書込が失敗する D1
      const cause = new Error("D1_ERROR: unavailable");
      const { database, calls } = await createFakeLocalD1Binding({
        first: async () => {
          throw cause;
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: 書込する
      const got = repository[operation](command);

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it("throws_d1_error_without_leaking_values_when_d1_fails", async () => {
      // Given: bind 値を含む message で失敗する D1
      const { database } = await createFakeLocalD1Binding({
        first: async () => {
          throw new Error("failed for ep-secret");
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: 書込する
      const got = await repository[operation]({ ...command, episodeId: "ep-secret" }).catch(
        (error: unknown) => error,
      );

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("ep-secret");
      expect((got as D1Error).message).not.toContain("INSERT");
    });

    it("throws_d1_error_when_returning_yields_no_row", async () => {
      // Given: RETURNING が行を返さない D1（upsert の契約違反）
      const { database } = await createFakeLocalD1Binding({ first: async () => null });
      const repository = new D1ProgressRepository({ database });

      // When / Then: 勝ち側が分からないので D1Error
      await expect(repository[operation](command)).rejects.toBeInstanceOf(D1Error);
    });

    it.each([
      [
        "first_played_at is not a string",
        { [columns.firstPlayedAt]: 1, [columns.firstCompletedAt]: null },
      ],
      [
        "first_completed_at is neither null nor a string",
        { [columns.firstPlayedAt]: "2026-10-01T00:00:00.000Z", [columns.firstCompletedAt]: 1 },
      ],
    ])("throws_d1_error_when_returning_row_is_malformed_because_%s", async (_name, row) => {
      // Given: 列型が期待と違う RETURNING 行を返す D1
      const { database } = await createFakeLocalD1Binding({ first: async () => row });
      const repository = new D1ProgressRepository({ database });

      // When / Then: 契約型として返さず D1Error
      await expect(repository[operation](command)).rejects.toBeInstanceOf(D1Error);
    });

    it("throws_without_calling_d1_when_client_at_is_not_a_valid_timestamp", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: Route 側 schema の前提を破る clientAt で書込する
      const got = repository[operation]({ ...command, clientAt: "not-a-timestamp" });

      // Then: 書込開始前に RangeError で止まり、D1 は呼ばれない
      await expect(got).rejects.toBeInstanceOf(RangeError);
      expect(calls).toHaveLength(0);
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

    it("binds_utc_normalized_since_when_since_has_offset", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: +09:00 の since で pull する
      await repository.listUpdatedSince("2026-10-01T09:00:00+09:00");

      // Then: 比較可能な UTC 固定幅表現へ揃えて 1 query だけ送る
      expect(calls).toHaveLength(1);
      expect(calls[0]?.values).toEqual(["2026-10-01T00:00:00.000Z"]);
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

    it("throws_without_calling_d1_when_since_is_not_a_valid_timestamp", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: Route 側 schema の前提を破る since で pull する
      const got = repository.listUpdatedSince("not-a-timestamp");

      // Then: RangeError で止まり、D1 は呼ばれない
      await expect(got).rejects.toBeInstanceOf(RangeError);
      expect(calls).toHaveLength(0);
    });
  });
});
