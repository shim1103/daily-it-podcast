// @vitest-environment node
/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers D1 binding（実 SQLite）
 * Double: 本番 remote D1 は使わない（remoteBindings: false）
 *
 * 実 SQL の全列置換（新規 INSERT と既存行の置換）・pull の境界と並び・実 D1 の bind 数上限・
 * 表なし失敗に加え、行なしの条件付き挿入（行なしで書く・既存行で書かない・並行）、版つきの条件付き置換
 * （版の一致・不一致・行なし、同じ期待値の並行）、版の単調性と一意性、cursor 差分取得の書込順を所有する。
 * 勝敗の merge 規則は Application（merge-progress の sociable unit）が所有する。
 * binding 呼び出しの形・Error 写像・再試行なしは sociable unit が所有する。
 *
 * @require createLocalD1Binding が実 proxy を起動し、applyEpisodeProgressMigration で表を作れる
 * @ensure D1ProgressRepository が実 SQLite に対して Port の永続契約（渡された行をそのまま保存し、読み返せる。行が無い時だけ挿入し、期待した版と一致する時だけ置き換える）を満たす
 */
import { afterAll, beforeAll, beforeEach, describe, expect, it } from "vitest";
import type {
  ProgressCursor,
  ProgressRow,
  ProgressVersion,
} from "../../worker/src/application/ports/progress-repository.ts";
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

// why: 契約の起点 `PROGRESS_PULL_ORIGIN_CURSOR` は文字列。adapter は変換しないので、Port の数値 cursor の起点をここで持つ
const originCursor: ProgressCursor = 0;

function progressRow(episodeId: string, overrides: Partial<ProgressRow> = {}): ProgressRow {
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
  return (await repository.getByEpisodeIds({ episodeIds: [episodeId] })).get(episodeId);
}

async function versionOf(episodeId: string): Promise<ProgressVersion> {
  const got = await repository.getProgressWithVersion({ episodeId });
  if (got === null) {
    throw new Error(`${episodeId} の行が無い`);
  }
  return got.version;
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
      const got = await repository.getByEpisodeIds({ episodeIds: ["ep-1", "ep-unknown"] });

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
      const got = await repository.getByEpisodeIds({ episodeIds });

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

  describe("getProgressWithVersion", () => {
    it("returns_null_when_row_does_not_exist", async () => {
      // Given: 行の無い D1（beforeEach で空）

      // When: 版つきで取得する
      const got = await repository.getProgressWithVersion({ episodeId: "ep-unknown" });

      // Then: null
      expect(got).toBeNull();
    });

    it("returns_progress_as_written_with_version_when_row_exists", async () => {
      // Given: 条件付き挿入で作った行
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 33 }) });

      // When: 版つきで取得する
      const got = await repository.getProgressWithVersion({ episodeId: "ep-1" });

      // Then: 書いた値のままの progress と、版が返る
      expect(got?.progress).toEqual({
        positionSec: 33,
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:10.000Z",
      });
      expect(got?.version).toEqual(expect.any(Number));
    });
  });

  describe("insertProgressIfAbsent", () => {
    it("returns_written_and_stores_row_when_no_row_exists", async () => {
      // Given: 行の無い episode

      // When: 条件付き挿入する
      const got = await repository.insertProgressIfAbsent({
        row: progressRow("ep-1", {
          positionSec: 60.5,
          firstCompletedAt: "2026-10-01T00:01:00.000Z",
        }),
      });

      // Then: "written" で、渡した値のまま読み返せる
      expect(got).toBe("written");
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 60.5,
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
        lastPlayedAt: "2026-10-01T00:00:10.000Z",
      });
    });

    it("returns_conflict_and_keeps_row_and_version_when_row_exists", async () => {
      // Given: 既に行がある episode
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 1 }) });
      const before = await repository.getProgressWithVersion({ episodeId: "ep-1" });

      // When: 行なしと見て条件付き挿入する
      const got = await repository.insertProgressIfAbsent({
        row: progressRow("ep-1", { positionSec: 99 }),
      });

      // Then: "conflict" で、行も版も変わらない
      expect(got).toBe("conflict");
      expect(await repository.getProgressWithVersion({ episodeId: "ep-1" })).toEqual(before);
    });

    it("returns_exactly_one_written_when_inserts_for_the_same_episode_run_concurrently", async () => {
      // Given: 行の無い episode
      const positions = Array.from({ length: 10 }, (_, index) => index + 1);

      // When: 同じ episode への条件付き挿入を 10 本並行して送る
      const results = await Promise.all(
        positions.map((positionSec) =>
          repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec }) }),
        ),
      );

      // Then: 勝者はちょうど 1 本で、行は勝者の値になる
      const winnerIndex = results.indexOf("written");
      expect(results.filter((result) => result === "written")).toHaveLength(1);
      expect(results.filter((result) => result === "conflict")).toHaveLength(positions.length - 1);
      expect((await progressOf("ep-1"))?.positionSec).toBe(positions[winnerIndex]);
    });
  });

  describe("replaceProgressIfVersion", () => {
    it("returns_written_and_replaces_values_when_expected_version_matches", async () => {
      // Given: 再生中の行と、その版
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 10 }) });
      const expectedVersion = await versionOf("ep-1");

      // When: その版を期待値に、再生が進んだ row（位置と最終再生が進む）で置き換える
      const got = await repository.replaceProgressIfVersion({
        row: progressRow("ep-1", {
          positionSec: 30,
          lastPlayedAt: "2026-10-01T00:00:30.000Z",
        }),
        expectedVersion,
      });

      // Then: "written" で、渡した値に置き換わる
      expect(got).toBe("written");
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 30,
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:30.000Z",
      });
    });

    it("stores_given_row_as_is_without_merging_when_expected_version_matches", async () => {
      // Given: 完走済みで first が古い行と、その版
      await repository.insertProgressIfAbsent({
        row: progressRow("ep-1", {
          positionSec: 60,
          firstPlayedAt: "2026-10-01T00:00:10.000Z",
          firstCompletedAt: "2026-10-01T00:01:00.000Z",
          lastPlayedAt: "2026-10-01T00:02:00.000Z",
        }),
      });
      const expectedVersion = await versionOf("ep-1");

      // why: Application が作らない入力を、adapter が勝手に merge していないことを検出するために与える（Decision 2026-10-01T18-54-16、永続は薄く保つ）
      // When: merge と食い違う row（未完走・first が新しい・最終再生が古い）で、版一致のまま置き換える
      await repository.replaceProgressIfVersion({
        row: progressRow("ep-1", {
          positionSec: 5,
          firstPlayedAt: "2026-10-01T00:00:20.000Z",
          firstCompletedAt: null,
          lastPlayedAt: "2026-10-01T00:00:30.000Z",
        }),
        expectedVersion,
      });

      // Then: 先勝ち・後勝ちの merge が掛からず、渡した値のまま全列が保存される（firstCompletedAt も null に戻る）
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 5,
        firstPlayedAt: "2026-10-01T00:00:20.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:30.000Z",
      });
    });

    it("returns_conflict_and_writes_nothing_when_expected_version_is_stale", async () => {
      // Given: 版を読んだ後に、別の書込で行が進んだ episode
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 1 }) });
      const staleVersion = await versionOf("ep-1");
      await repository.replaceProgressIfVersion({
        row: progressRow("ep-1", { positionSec: 2 }),
        expectedVersion: staleVersion,
      });
      const before = await repository.getProgressWithVersion({ episodeId: "ep-1" });

      // When: 古い版を期待値に置き換える
      const got = await repository.replaceProgressIfVersion({
        row: progressRow("ep-1", { positionSec: 99 }),
        expectedVersion: staleVersion,
      });

      // Then: "conflict" で、行も版も変わらない
      expect(got).toBe("conflict");
      expect(await repository.getProgressWithVersion({ episodeId: "ep-1" })).toEqual(before);
    });

    it("returns_conflict_and_inserts_nothing_when_row_does_not_exist", async () => {
      // Given: 版を読んだ後に、行が消えた episode
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1") });
      const expectedVersion = await versionOf("ep-1");
      await database.prepare(`DELETE FROM ${EPISODE_PROGRESS_TABLE}`).run();

      // When: 消えた行の版を期待値に置き換える
      const got = await repository.replaceProgressIfVersion({
        row: progressRow("ep-1"),
        expectedVersion,
      });

      // Then: "conflict" で、行は作られない
      expect(got).toBe("conflict");
      expect(await progressOf("ep-1")).toBeUndefined();
    });

    it("leaves_other_episode_rows_untouched_when_one_row_is_replaced", async () => {
      // Given: 2 episode の行
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 1 }) });
      await repository.insertProgressIfAbsent({ row: progressRow("ep-2", { positionSec: 2 }) });
      const otherBefore = await repository.getProgressWithVersion({ episodeId: "ep-2" });

      // When: ep-1 だけ版一致で置き換える
      await repository.replaceProgressIfVersion({
        row: progressRow("ep-1", { positionSec: 99 }),
        expectedVersion: await versionOf("ep-1"),
      });

      // Then: ep-2 の行と版は変わらない
      expect(await repository.getProgressWithVersion({ episodeId: "ep-2" })).toEqual(otherBefore);
    });

    it("returns_exactly_one_written_when_replaces_share_the_same_expected_version_concurrently", async () => {
      // Given: 1 行と、その版
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 0 }) });
      const expectedVersion = await versionOf("ep-1");
      const positions = Array.from({ length: 10 }, (_, index) => index + 1);

      // When: 同じ期待値で 10 本の置き換えを並行して送る
      const results = await Promise.all(
        positions.map((positionSec) =>
          repository.replaceProgressIfVersion({
            row: progressRow("ep-1", { positionSec }),
            expectedVersion,
          }),
        ),
      );

      // Then: 勝者はちょうど 1 本で、行は勝者の値になる
      const winnerIndex = results.indexOf("written");
      expect(results.filter((result) => result === "written")).toHaveLength(1);
      expect(results.filter((result) => result === "conflict")).toHaveLength(positions.length - 1);
      expect((await progressOf("ep-1"))?.positionSec).toBe(positions[winnerIndex]);
    });
  });

  describe("version numbering", () => {
    it("advances_version_monotonically_and_uniquely_across_rows_when_rows_are_written", async () => {
      // Given: 空の D1

      // When: 2 episode の挿入と、ep-1 の置き換えを順に行い、written のたびに版を読む
      const versions: ProgressVersion[] = [];
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1") });
      versions.push(await versionOf("ep-1"));
      await repository.insertProgressIfAbsent({ row: progressRow("ep-2") });
      versions.push(await versionOf("ep-2"));
      await repository.replaceProgressIfVersion({
        row: progressRow("ep-1", { positionSec: 20 }),
        expectedVersion: versions[0] as ProgressVersion,
      });
      versions.push(await versionOf("ep-1"));

      // Then: 版は書込のたびに厳密に増え、全て異なる
      expect(
        versions.every(
          (version, index) => index === 0 || version > (versions[index - 1] as number),
        ),
      ).toBe(true);
      expect(new Set(versions).size).toBe(versions.length);
    });

    it("assigns_distinct_versions_when_different_episodes_are_inserted_concurrently", async () => {
      // Given: 行の無い 10 episode
      const episodeIds = Array.from({ length: 10 }, (_, index) => `ep-${index}`);

      // When: 別 episode の条件付き挿入を 10 本並行して送る
      const results = await Promise.all(
        episodeIds.map((episodeId) =>
          repository.insertProgressIfAbsent({ row: progressRow(episodeId) }),
        ),
      );

      // Then: 全て written で、版は行ごとに異なる
      expect(results.every((result) => result === "written")).toBe(true);
      const versions = await Promise.all(episodeIds.map((episodeId) => versionOf(episodeId)));
      expect(new Set(versions).size).toBe(episodeIds.length);
    });
  });

  describe("listChangedAfter", () => {
    it("returns_same_cursor_and_empty_entries_when_database_is_empty", async () => {
      // Given: 空の D1

      // When: 起点の cursor で差分取得する
      const got = await repository.listChangedAfter({ cursor: originCursor });

      // Then: 引数の cursor を数値のまま返し、entries は空配列
      expect(got).toEqual({ cursor: originCursor, entries: [] });
    });

    it("returns_rows_in_write_order_with_cursor_of_last_row_when_rows_are_written", async () => {
      // Given: ep-b、ep-a の順に挿入し、ep-b をもう一度置き換えた D1（episodeId 順・初回書込順とは別の並びになる）
      await repository.insertProgressIfAbsent({ row: progressRow("ep-b") });
      await repository.insertProgressIfAbsent({ row: progressRow("ep-a") });
      await repository.replaceProgressIfVersion({
        row: progressRow("ep-b", { positionSec: 77 }),
        expectedVersion: await versionOf("ep-b"),
      });

      // When: 起点から差分取得する
      const got = await repository.listChangedAfter({ cursor: originCursor });

      // Then: 最後に書いた順（ep-a、ep-b）で返り、cursor は最後の行の版と同じ数値になる
      expect(got.entries.map((entry) => entry.episodeId)).toEqual(["ep-a", "ep-b"]);
      expect(got.entries[1]?.progress.positionSec).toBe(77);
      expect(got.cursor).toBe(await versionOf("ep-b"));
    });

    it("returns_only_rows_written_after_cursor_when_cursor_is_returned_value", async () => {
      // Given: 1 回目の差分取得で cursor を受け取った後に、別 episode が書かれた D1
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1") });
      const first = await repository.listChangedAfter({ cursor: originCursor });
      await repository.insertProgressIfAbsent({ row: progressRow("ep-2") });

      // When: 受け取った cursor で差分取得する
      const got = await repository.listChangedAfter({ cursor: first.cursor });

      // Then: 後から書かれた行だけが返る
      expect(got.entries.map((entry) => entry.episodeId)).toEqual(["ep-2"]);
    });

    it("returns_same_cursor_and_empty_entries_when_nothing_changed_after_cursor", async () => {
      // Given: 差分取得で cursor を受け取った後、何も書かれていない D1
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1") });
      const first = await repository.listChangedAfter({ cursor: originCursor });

      // When: 受け取った cursor で差分取得する
      const got = await repository.listChangedAfter({ cursor: first.cursor });

      // Then: 空配列と、同じ cursor
      expect(got).toEqual({ cursor: first.cursor, entries: [] });
    });

    it("returns_no_row_for_conflicting_write_when_write_is_rejected", async () => {
      // Given: 差分取得で cursor を受け取った後、競合で弾かれた書込がある D1
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1") });
      const first = await repository.listChangedAfter({ cursor: originCursor });
      await repository.insertProgressIfAbsent({ row: progressRow("ep-1", { positionSec: 99 }) });

      // When: 受け取った cursor で差分取得する
      const got = await repository.listChangedAfter({ cursor: first.cursor });

      // Then: 弾かれた書込は cursor を進めず、行にも現れない
      expect(got).toEqual({ cursor: first.cursor, entries: [] });
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
