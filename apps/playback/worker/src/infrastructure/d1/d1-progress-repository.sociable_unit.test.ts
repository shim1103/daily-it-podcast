import { describe, expect, it } from "vitest";
import { createFakeLocalD1Binding } from "../../../../test/support/create-fake-local-d1-binding.ts";
import type { ProgressRow, ProgressVersion } from "../../application/ports/progress-repository.ts";
import type { D1Row } from "./d1-database-binding.ts";
import { D1Error } from "./d1-error.ts";
import { D1ProgressRepository } from "./d1-progress-repository.ts";
import {
  D1_MAX_BOUND_PARAMETERS_PER_QUERY,
  EPISODE_PROGRESS_TABLE,
  PROGRESS_D1_ADAPTER_MAX_ATTEMPTS,
  episodeProgressColumns,
} from "./progress-d1-constants.ts";

/**
 * scope: Sociable Unit
 * real: D1ProgressRepository
 * double: D1DatabaseBinding（共有 Fake local binding。応答を差し込み、呼び出しを記録する）
 *
 * 本 file は、Port 各 method が送る SQL の形・bind 値と順序・`changes` から結果への写像・失敗写像・cursor を数値のまま
 * 受け渡すことを所有する。勝敗の merge 規則は Application（`merge-progress.sociable_unit.test.ts`）が所有する。
 * 全列置換・pull の境界と並び・bind 数上限・条件付き書込の勝敗と採番の実 SQLite での意味は
 * `test/integration/d1_progress_repository.narrow_integration.test.ts` が所有する。
 */

const columns = episodeProgressColumns;

// 採番の取り方（現在の最大値 + 1。空表は 1 から）
const nextSeqSubquery = `(SELECT COALESCE(MAX(${columns.seq}), 0) + 1 FROM ${EPISODE_PROGRESS_TABLE})`;

const writeRow: ProgressRow = {
  episodeId: "ep-1",
  positionSec: 42.5,
  firstPlayedAt: "2026-10-01T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-10-01T00:05:00.000Z",
};

// why: ProgressVersion は Port が公開する不透明な型で、生成手段は adapter が読んで返す経路だけ。入力値を作る test 用に限り brand を付ける
function versionOf(value: number): ProgressVersion {
  return value as ProgressVersion;
}

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
      const got = await repository.getByEpisodeIds({ episodeIds: ["ep-1", "ep-2"] });

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
      const got = await repository.getByEpisodeIds({ episodeIds: ["ep-1"] });

      // Then: 空 Map（null でない）
      expect(got.size).toBe(0);
    });

    it("returns_empty_map_without_calling_d1_when_episode_ids_is_empty", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 空配列で取得する
      const got = await repository.getByEpisodeIds({ episodeIds: [] });

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
      const got = await repository.getByEpisodeIds({ episodeIds });

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
      await repository.getByEpisodeIds({ episodeIds });

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
      await repository.getByEpisodeIds({ episodeIds });

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
      const got = repository.getByEpisodeIds({ episodeIds: ["ep-1"] });

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
      const got = await repository
        .getByEpisodeIds({ episodeIds: ["ep-secret"] })
        .catch((error: unknown) => error);

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

  describe("getProgressWithVersion", () => {
    it("returns_progress_with_version_when_row_exists", async () => {
      // Given: ep-1 の行が採番 7 で存在する D1
      const { database } = await createFakeLocalD1Binding({
        first: async () => progressRow({ [columns.seq]: 7 }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 版つきで取得する
      const got = await repository.getProgressWithVersion({ episodeId: "ep-1" });

      // Then: 列が契約型の progress へ写り、版は不透明な値として添う。採番の列名は応答に出ない
      expect(got).toEqual({
        progress: {
          positionSec: 12.5,
          firstPlayedAt: "2026-10-01T00:00:00.000Z",
          firstCompletedAt: null,
          lastPlayedAt: "2026-10-01T00:05:00.000Z",
        },
        version: 7,
      });
    });

    it("returns_null_when_row_does_not_exist", async () => {
      // Given: 行が 1 つも無い D1（Fake の既定応答）
      const { database } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 版つきで取得する
      const got = await repository.getProgressWithVersion({ episodeId: "ep-unknown" });

      // Then: null（throw でない）
      expect(got).toBeNull();
    });

    it("selects_one_row_by_episode_id_with_seq_column_in_one_query", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 版つきで取得する
      await repository.getProgressWithVersion({ episodeId: "ep-1" });

      // Then: 1 query だけ送り、採番の列を選び、episodeId を値のまま bind して鍵で絞る
      expect(calls).toHaveLength(1);
      expect(calls[0]?.sql).toContain(columns.seq);
      expect(calls[0]?.sql).toContain(
        `FROM ${EPISODE_PROGRESS_TABLE} WHERE ${columns.episodeId} = ?1`,
      );
      expect(calls[0]?.values).toEqual(["ep-1"]);
    });

    it("throws_d1_error_with_cause_without_retry_when_d1_fails", async () => {
      // Given: 読取が失敗する D1
      const cause = new Error("D1_ERROR: unavailable");
      const { database, calls } = await createFakeLocalD1Binding({
        first: async () => {
          throw cause;
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: 版つきで取得する
      const got = repository.getProgressWithVersion({ episodeId: "ep-1" });

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

      // When: 版つきで取得する
      const got = await repository
        .getProgressWithVersion({ episodeId: "ep-secret" })
        .catch((error: unknown) => error);

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("ep-secret");
      expect((got as D1Error).message).not.toContain("SELECT");
    });

    it.each([
      ["seq is not a number", { [columns.seq]: "7" }],
      ["seq is null", { [columns.seq]: null }],
      ["position_sec is not a number", { [columns.positionSec]: "12" }],
    ])("throws_d1_error_when_row_is_malformed_because_%s", async (_name, overrides) => {
      // Given: 列型が期待と違う行を返す D1
      const { database } = await createFakeLocalD1Binding({
        first: async () => progressRow({ [columns.seq]: 7, ...overrides }),
      });
      const repository = new D1ProgressRepository({ database });

      // When / Then: 版や契約型として返さず D1Error
      await expect(repository.getProgressWithVersion({ episodeId: "ep-1" })).rejects.toBeInstanceOf(
        D1Error,
      );
    });
  });

  describe("insertProgressIfAbsent", () => {
    it("returns_written_when_one_row_is_inserted", async () => {
      // Given: 1 行の挿入を報告する D1
      const { database } = await createFakeLocalD1Binding({
        run: async () => ({ success: true, meta: { changes: 1 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 行なしと見て条件付き挿入する
      const got = await repository.insertProgressIfAbsent({ row: writeRow });

      // Then: "written"
      expect(got).toBe("written");
    });

    it("returns_conflict_when_no_row_is_inserted", async () => {
      // Given: 挿入が 0 行（既に行がある）と報告する D1
      const { database } = await createFakeLocalD1Binding({
        run: async () => ({ success: true, meta: { changes: 0 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 条件付き挿入する
      const got = await repository.insertProgressIfAbsent({ row: writeRow });

      // Then: throw せず "conflict"
      expect(got).toBe("conflict");
    });

    it("sends_one_insert_that_does_nothing_on_conflict_with_seq_numbered_in_same_statement", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 条件付き挿入する
      await repository.insertProgressIfAbsent({ row: writeRow });

      // Then: 読まずに 1 文だけ送り、既存行は書き換えず、採番は同じ文の中の subquery で行う
      const sql = calls[0]?.sql ?? "";
      expect(calls).toHaveLength(1);
      expect(sql).toContain(`INSERT INTO ${EPISODE_PROGRESS_TABLE}`);
      expect(sql).toContain(`ON CONFLICT(${columns.episodeId}) DO NOTHING`);
      expect(sql).toContain(nextSeqSubquery);
      expect(sql).not.toContain("DO UPDATE");
      expect(sql).not.toContain("UPDATE SET");
    });

    it("binds_row_columns_in_table_order_without_seq", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 完走済みの row を条件付き挿入する
      await repository.insertProgressIfAbsent({
        row: { ...writeRow, firstCompletedAt: "2026-10-01T00:04:00.000Z" },
      });

      // Then: row の列を値のまま表の列順に bind する。採番は SQL が持つので bind しない
      expect(calls[0]?.values).toEqual([
        "ep-1",
        42.5,
        "2026-10-01T00:00:00.000Z",
        "2026-10-01T00:04:00.000Z",
        "2026-10-01T00:05:00.000Z",
      ]);
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

      // When: 条件付き挿入する
      const got = repository.insertProgressIfAbsent({ row: writeRow });

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it("throws_d1_error_without_retry_when_d1_reports_unsuccessful_run", async () => {
      // Given: success が false の結果を返す D1（changes は 1 でも成功扱いにしない）
      const { database, calls } = await createFakeLocalD1Binding({
        run: async () => ({ success: false, meta: { changes: 1 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 条件付き挿入する
      const got = repository.insertProgressIfAbsent({ row: writeRow });

      // Then: "written" を返さず D1Error を throw し、試行は上限（再試行なし）に収まる
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
        { run: async () => ({ success: false, meta: { changes: 1 } }) },
      ],
    ])("throws_d1_error_without_leaking_values_when_%s", async (_name, responses) => {
      // Given: 失敗する D1
      const { database } = await createFakeLocalD1Binding(responses);
      const repository = new D1ProgressRepository({ database });

      // When: bind 値を含む row を条件付き挿入する
      const got = await repository
        .insertProgressIfAbsent({ row: { ...writeRow, episodeId: "ep-secret" } })
        .catch((error: unknown) => error);

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("ep-secret");
      expect((got as D1Error).message).not.toContain("INSERT");
    });
  });

  describe("replaceProgressIfVersion", () => {
    it("returns_written_when_one_row_is_updated", async () => {
      // Given: 1 行の更新を報告する D1
      const { database } = await createFakeLocalD1Binding({
        run: async () => ({ success: true, meta: { changes: 1 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 期待値 3 で条件付き更新する
      const got = await repository.replaceProgressIfVersion({
        row: writeRow,
        expectedVersion: versionOf(3),
      });

      // Then: "written"
      expect(got).toBe("written");
    });

    it("returns_conflict_when_no_row_is_updated", async () => {
      // Given: 更新が 0 行（版が食い違う、または行が無い）と報告する D1
      const { database } = await createFakeLocalD1Binding({
        run: async () => ({ success: true, meta: { changes: 0 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 期待値 3 で条件付き更新する
      const got = await repository.replaceProgressIfVersion({
        row: writeRow,
        expectedVersion: versionOf(3),
      });

      // Then: throw せず "conflict"
      expect(got).toBe("conflict");
    });

    it("sends_one_update_guarded_by_episode_id_and_expected_seq_with_seq_numbered_in_same_statement", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 期待値 3 で条件付き更新する
      await repository.replaceProgressIfVersion({ row: writeRow, expectedVersion: versionOf(3) });

      // Then: 読まずに 1 文だけ送り、鍵以外の 4 列と採番を更新し、鍵と期待値が一致する行だけを対象にする
      const sql = calls[0]?.sql ?? "";
      expect(calls).toHaveLength(1);
      expect(sql).toContain(`UPDATE ${EPISODE_PROGRESS_TABLE} SET`);
      expect(sql).toContain(`${columns.positionSec} = ?2`);
      expect(sql).toContain(`${columns.firstPlayedAt} = ?3`);
      expect(sql).toContain(`${columns.firstCompletedAt} = ?4`);
      expect(sql).toContain(`${columns.lastPlayedAt} = ?5`);
      expect(sql).toContain(`${columns.seq} = ${nextSeqSubquery}`);
      expect(sql).toContain(`WHERE ${columns.episodeId} = ?1 AND ${columns.seq} = ?6`);
      expect(sql).not.toContain("INSERT");
    });

    it("binds_row_columns_in_table_order_then_expected_version_as_number", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: 完走済みの row を期待値 3 で条件付き更新する
      await repository.replaceProgressIfVersion({
        row: { ...writeRow, firstCompletedAt: "2026-10-01T00:04:00.000Z" },
        expectedVersion: versionOf(3),
      });

      // Then: row の列を値のまま表の列順に bind し、末尾に期待値を数値で bind する
      expect(calls[0]?.values).toEqual([
        "ep-1",
        42.5,
        "2026-10-01T00:00:00.000Z",
        "2026-10-01T00:04:00.000Z",
        "2026-10-01T00:05:00.000Z",
        3,
      ]);
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

      // When: 条件付き更新する
      const got = repository.replaceProgressIfVersion({
        row: writeRow,
        expectedVersion: versionOf(3),
      });

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it("throws_d1_error_without_retry_when_d1_reports_unsuccessful_run", async () => {
      // Given: success が false の結果を返す D1（changes は 1 でも成功扱いにしない）
      const { database, calls } = await createFakeLocalD1Binding({
        run: async () => ({ success: false, meta: { changes: 1 } }),
      });
      const repository = new D1ProgressRepository({ database });

      // When: 条件付き更新する
      const got = repository.replaceProgressIfVersion({
        row: writeRow,
        expectedVersion: versionOf(3),
      });

      // Then: "written" を返さず D1Error を throw し、試行は上限（再試行なし）に収まる
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
        { run: async () => ({ success: false, meta: { changes: 1 } }) },
      ],
    ])("throws_d1_error_without_leaking_values_when_%s", async (_name, responses) => {
      // Given: 失敗する D1
      const { database } = await createFakeLocalD1Binding(responses);
      const repository = new D1ProgressRepository({ database });

      // When: bind 値を含む row を条件付き更新する
      const got = await repository
        .replaceProgressIfVersion({
          row: { ...writeRow, episodeId: "ep-secret" },
          expectedVersion: versionOf(3),
        })
        .catch((error: unknown) => error);

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("ep-secret");
      expect((got as D1Error).message).not.toContain("UPDATE");
    });
  });

  describe("listChangedAfter", () => {
    it("returns_entries_in_row_order_with_numeric_cursor_of_last_row_seq", async () => {
      // Given: 採番 5 と 9 の 2 行を返す D1
      const { database } = await createFakeLocalD1Binding({
        all: async () => [
          progressRow({ [columns.episodeId]: "ep-1", [columns.seq]: 5 }),
          progressRow({
            [columns.episodeId]: "ep-2",
            [columns.seq]: 9,
            [columns.positionSec]: 3,
            [columns.firstCompletedAt]: "2026-10-01T00:06:00.000Z",
          }),
        ],
      });
      const repository = new D1ProgressRepository({ database });

      // When: cursor 4 で差分取得する
      const got = await repository.listChangedAfter({ cursor: 4 });

      // Then: 行の並びのまま契約型へ写り（採番は載せない）、cursor は最後の行の採番を数値のまま返す
      expect(got).toEqual({
        cursor: 9,
        entries: [
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
        ],
      });
    });

    it("returns_same_numeric_cursor_and_empty_entries_when_no_row_is_after_cursor", async () => {
      // Given: 行が 1 つも無い D1（Fake の既定応答）
      const { database } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: cursor 7 で差分取得する
      const got = await repository.listChangedAfter({ cursor: 7 });

      // Then: 引数の cursor を数値のまま返し、entries は空配列（null でない）
      expect(got).toEqual({ cursor: 7, entries: [] });
    });

    it("selects_rows_after_cursor_in_seq_order_with_cursor_bound_as_given", async () => {
      // Given: 呼び出しを記録する D1
      const { database, calls } = await createFakeLocalD1Binding();
      const repository = new D1ProgressRepository({ database });

      // When: cursor 7 で差分取得する
      await repository.listChangedAfter({ cursor: 7 });

      // Then: 1 query だけ送り、採番が cursor より大きい行を採番の昇順で選ぶ。cursor は数値のまま bind する
      expect(calls).toHaveLength(1);
      expect(calls[0]?.sql).toContain(`WHERE ${columns.seq} > ?1 ORDER BY ${columns.seq} ASC`);
      expect(calls[0]?.values).toEqual([7]);
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

      // When: 差分取得する
      const got = repository.listChangedAfter({ cursor: 7 });

      // Then: D1Error を cause 付きで throw し、試行は上限（再試行なし）に収まる
      await expect(got).rejects.toBeInstanceOf(D1Error);
      await expect(got).rejects.toHaveProperty("cause", cause);
      expect(calls).toHaveLength(PROGRESS_D1_ADAPTER_MAX_ATTEMPTS);
    });

    it("throws_d1_error_without_leaking_cursor_when_d1_fails", async () => {
      // Given: bind 値を含む message で失敗する D1
      const { database } = await createFakeLocalD1Binding({
        all: async () => {
          throw new Error("failed for cursor 4242");
        },
      });
      const repository = new D1ProgressRepository({ database });

      // When: 差分取得する
      const got = await repository
        .listChangedAfter({ cursor: 4242 })
        .catch((error: unknown) => error);

      // Then: D1Error の message に SQL 全文も bind 値も含まない
      expect(got).toBeInstanceOf(D1Error);
      expect((got as D1Error).message).not.toContain("4242");
      expect((got as D1Error).message).not.toContain("SELECT");
    });

    it.each([
      ["seq is not a number", { [columns.seq]: "9" }],
      ["seq is null", { [columns.seq]: null }],
      ["position_sec is not a number", { [columns.seq]: 9, [columns.positionSec]: "12" }],
    ])("throws_d1_error_when_row_is_malformed_because_%s", async (_name, overrides) => {
      // Given: 列型が期待と違う行を返す D1
      const { database } = await createFakeLocalD1Binding({
        all: async () => [progressRow(overrides)],
      });
      const repository = new D1ProgressRepository({ database });

      // When / Then: cursor や契約型として返さず D1Error
      await expect(repository.listChangedAfter({ cursor: 7 })).rejects.toBeInstanceOf(D1Error);
    });
  });
});
