import { describe, expect, it } from "vitest";
import type {
  ProgressCursor,
  ProgressRow,
  ProgressVersion,
} from "../../application/ports/progress-repository.ts";
import { InMemoryProgressRepository } from "./in-memory-progress-repository.ts";

/**
 * scope: Sociable Unit
 * real: InMemoryProgressRepository
 * double: なし
 */
const rowA: ProgressRow = {
  episodeId: "ep-a",
  positionSec: 10,
  firstPlayedAt: "2026-09-01T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-09-10T12:00:00.000Z",
};

const rowB: ProgressRow = {
  episodeId: "ep-b",
  positionSec: 20,
  firstPlayedAt: "2026-09-02T00:00:00.000Z",
  firstCompletedAt: "2026-09-05T00:00:00.000Z",
  lastPlayedAt: "2026-09-20T12:00:00.000Z",
};

const replacedA: ProgressRow = {
  episodeId: "ep-a",
  positionSec: 99,
  firstPlayedAt: "2026-09-15T00:00:00.000Z",
  firstCompletedAt: "2026-09-16T00:00:00.000Z",
  lastPlayedAt: "2026-09-17T00:00:00.000Z",
};

const intermediateA: ProgressRow = {
  episodeId: "ep-a",
  positionSec: 50,
  firstPlayedAt: "2026-09-03T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-09-12T00:00:00.000Z",
};

// why: 契約の起点 `PROGRESS_PULL_ORIGIN_CURSOR` は文字列。infrastructure は contracts を import しないので、Port の数値 cursor の起点をここで持つ
const originCursor: ProgressCursor = 0;

async function readVersion(
  repository: InMemoryProgressRepository,
  episodeId: string,
): Promise<ProgressVersion> {
  const got = await repository.getProgressWithVersion({ episodeId });
  if (got === null) {
    throw new Error(`行が無い episodeId の版は読めない: ${episodeId}`);
  }
  return got.version;
}

describe("InMemoryProgressRepository", () => {
  it("getByEpisodeIds_returns_seeded_rows_and_omits_missing_ids", async () => {
    // Given: seed 済みの repository
    const repository = new InMemoryProgressRepository([rowA, rowB]);

    // When: 存在する id と存在しない id を混ぜて取得する
    const got = await repository.getByEpisodeIds({ episodeIds: ["ep-a", "ep-missing", "ep-b"] });

    // Then: seed 行だけが入り、欠けた id は載らない
    expect(got.size).toBe(2);
    expect(got.get("ep-a")).toEqual({
      positionSec: 10,
      firstPlayedAt: "2026-09-01T00:00:00.000Z",
      firstCompletedAt: null,
      lastPlayedAt: "2026-09-10T12:00:00.000Z",
    });
    expect(got.get("ep-b")).toEqual({
      positionSec: 20,
      firstPlayedAt: "2026-09-02T00:00:00.000Z",
      firstCompletedAt: "2026-09-05T00:00:00.000Z",
      lastPlayedAt: "2026-09-20T12:00:00.000Z",
    });
    expect(got.has("ep-missing")).toBe(false);
  });

  it("upsertProgress_replaces_existing_row_without_merging", async () => {
    // Given: 空の repository
    const repository = new InMemoryProgressRepository();

    // When: 同じ episodeId を 2 回 upsert する（2 回目は全 field 差し替え）
    await repository.upsertProgress(rowA);
    const replaced: ProgressRow = {
      episodeId: "ep-a",
      positionSec: 99,
      firstPlayedAt: "2026-09-15T00:00:00.000Z",
      firstCompletedAt: "2026-09-16T00:00:00.000Z",
      lastPlayedAt: "2026-09-17T00:00:00.000Z",
    };
    await repository.upsertProgress(replaced);
    const got = await repository.getByEpisodeIds({ episodeIds: ["ep-a"] });

    // Then: 2 回目の行がそのまま残り、1 回目の値は混ざらない
    expect(got.get("ep-a")).toEqual({
      positionSec: 99,
      firstPlayedAt: "2026-09-15T00:00:00.000Z",
      firstCompletedAt: "2026-09-16T00:00:00.000Z",
      lastPlayedAt: "2026-09-17T00:00:00.000Z",
    });
  });

  it("listUpdatedSince_returns_only_rows_with_last_played_at_after_since", async () => {
    // Given: lastPlayedAt が前後で分かれる 2 行
    const repository = new InMemoryProgressRepository([rowA, rowB]);

    // When: rowA と rowB の間の since で差分取得する
    const got = await repository.listUpdatedSince("2026-09-15T00:00:00.000Z");

    // Then: rowB のみ。形は pull 契約の episodes 要素と同形
    expect(got).toEqual([
      {
        episodeId: "ep-b",
        progress: {
          positionSec: 20,
          firstPlayedAt: "2026-09-02T00:00:00.000Z",
          firstCompletedAt: "2026-09-05T00:00:00.000Z",
          lastPlayedAt: "2026-09-20T12:00:00.000Z",
        },
      },
    ]);
  });

  it("listUpdatedSince_excludes_row_with_last_played_at_equal_to_since", async () => {
    // Given: lastPlayedAt が since と同一の行
    const repository = new InMemoryProgressRepository([rowA]);

    // When: rowA.lastPlayedAt と同じ since で差分取得する
    const got = await repository.listUpdatedSince(rowA.lastPlayedAt);

    // Then: 空（境界は厳密な >）
    expect(got).toEqual([]);
  });

  describe("getProgressWithVersion", () => {
    it("returns_progress_with_version_when_row_is_seeded", async () => {
      // Given: ep-a、ep-b の順で seed し、後に seed した ep-b の版を読んだ repository
      const repository = new InMemoryProgressRepository([rowA, rowB]);
      const versionOfLaterSeed = await readVersion(repository, "ep-b");

      // When: ep-a を版つきで取得する
      const got = await repository.getProgressWithVersion({ episodeId: "ep-a" });

      // Then: 進捗本体が返り、版は後に seed した行より小さい
      expect(got?.progress).toEqual({
        positionSec: 10,
        firstPlayedAt: "2026-09-01T00:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-09-10T12:00:00.000Z",
      });
      expect(got?.version).toBeLessThan(versionOfLaterSeed);
    });

    it("returns_null_when_row_is_missing", async () => {
      // Given: ep-a だけ seed した repository
      const repository = new InMemoryProgressRepository([rowA]);

      // When: 行の無い episodeId を版つきで取得する
      const got = await repository.getProgressWithVersion({ episodeId: "ep-missing" });

      // Then: null を返す
      expect(got).toBeNull();
    });
  });

  describe("insertProgressIfAbsent", () => {
    it("writes_row_and_returns_written_when_row_is_missing", async () => {
      // Given: 空の repository
      const repository = new InMemoryProgressRepository();

      // When: 行の無い ep-a を書く
      const got = await repository.insertProgressIfAbsent({ row: rowA });
      const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-a"] });

      // Then: written を返し、行が書かれる
      expect(got).toBe("written");
      expect(stored.get("ep-a")).toEqual({
        positionSec: 10,
        firstPlayedAt: "2026-09-01T00:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-09-10T12:00:00.000Z",
      });
    });

    it("returns_conflict_and_keeps_existing_row_when_row_exists", async () => {
      // Given: ep-a を seed し、その時点の版と cursor を読んだ repository
      const repository = new InMemoryProgressRepository([rowA]);
      const versionBeforeConflict = await readVersion(repository, "ep-a");
      const cursorBeforeConflict = (await repository.listChangedAfter({ cursor: originCursor }))
        .cursor;

      // When: 既に行がある ep-a へ別の値を書く
      const got = await repository.insertProgressIfAbsent({ row: replacedA });
      const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-a"] });
      const versionAfterConflict = await readVersion(repository, "ep-a");
      const changes = await repository.listChangedAfter({ cursor: originCursor });

      // Then: conflict を返し、既存の行の値・版と cursor は動かない
      expect(got).toBe("conflict");
      expect(stored.get("ep-a")).toEqual({
        positionSec: 10,
        firstPlayedAt: "2026-09-01T00:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-09-10T12:00:00.000Z",
      });
      expect(versionAfterConflict).toBe(versionBeforeConflict);
      expect(changes.cursor).toBe(cursorBeforeConflict);
    });
  });

  describe("replaceProgressIfVersion", () => {
    it("replaces_row_and_returns_written_when_expected_version_matches", async () => {
      // Given: ep-a を seed し、その時点の版を読んだ repository
      const repository = new InMemoryProgressRepository([rowA]);
      const version = await readVersion(repository, "ep-a");

      // When: 読んだ版を期待値にして別の値を書く
      const got = await repository.replaceProgressIfVersion({
        row: replacedA,
        expectedVersion: version,
      });
      const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-a"] });

      // Then: written を返し、行が全 field 置き換わる
      expect(got).toBe("written");
      expect(stored.get("ep-a")).toEqual({
        positionSec: 99,
        firstPlayedAt: "2026-09-15T00:00:00.000Z",
        firstCompletedAt: "2026-09-16T00:00:00.000Z",
        lastPlayedAt: "2026-09-17T00:00:00.000Z",
      });
    });

    it("returns_conflict_and_writes_nothing_when_expected_version_is_stale", async () => {
      // Given: 古い版を読んだ後に、seed と別の値で書込が入り、行の版が進んだ repository
      const repository = new InMemoryProgressRepository([rowA]);
      const staleVersion = await readVersion(repository, "ep-a");
      await repository.replaceProgressIfVersion({
        row: intermediateA,
        expectedVersion: staleVersion,
      });
      const versionBeforeConflict = await readVersion(repository, "ep-a");
      const cursorBeforeConflict = (await repository.listChangedAfter({ cursor: originCursor }))
        .cursor;

      // When: 古い版を期待値にして別の値を書く
      const got = await repository.replaceProgressIfVersion({
        row: replacedA,
        expectedVersion: staleVersion,
      });
      const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-a"] });
      const versionAfterConflict = await readVersion(repository, "ep-a");
      const changes = await repository.listChangedAfter({ cursor: originCursor });

      // Then: conflict を返し、行の値と版と cursor は途中の書込のまま動かない
      expect(got).toBe("conflict");
      expect(stored.get("ep-a")).toEqual({
        positionSec: 50,
        firstPlayedAt: "2026-09-03T00:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-09-12T00:00:00.000Z",
      });
      expect(versionAfterConflict).toBe(versionBeforeConflict);
      expect(changes.cursor).toBe(cursorBeforeConflict);
    });

    it("returns_conflict_and_writes_nothing_when_row_is_missing", async () => {
      // Given: ep-b だけがあり、その版と cursor を読んだ、ep-a の行が無い repository
      const repository = new InMemoryProgressRepository([rowB]);
      const otherRowVersion = await readVersion(repository, "ep-b");
      const cursorBeforeConflict = (await repository.listChangedAfter({ cursor: originCursor }))
        .cursor;

      // When: 行の無い ep-a へ、他の行の版を期待値にして書く
      const got = await repository.replaceProgressIfVersion({
        row: rowA,
        expectedVersion: otherRowVersion,
      });
      const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-a"] });
      const changes = await repository.listChangedAfter({ cursor: originCursor });

      // Then: conflict を返し、ep-a の行は作られず、cursor は動かない
      expect(got).toBe("conflict");
      expect(stored.size).toBe(0);
      expect(changes.cursor).toBe(cursorBeforeConflict);
    });
  });

  describe("version numbering", () => {
    it("advances_version_monotonically_and_keeps_it_unique_across_rows_when_written", async () => {
      // Given: 空の repository
      const repository = new InMemoryProgressRepository();

      // When: ep-a、ep-b の順に挿入し、ep-a を直前に読んだ版を期待値にして置き換える
      await repository.insertProgressIfAbsent({ row: rowA });
      const versionA1 = await readVersion(repository, "ep-a");
      await repository.insertProgressIfAbsent({ row: rowB });
      const versionB1 = await readVersion(repository, "ep-b");
      await repository.replaceProgressIfVersion({ row: replacedA, expectedVersion: versionA1 });
      const versionA2 = await readVersion(repository, "ep-a");

      // Then: 書込のたびに版が増え、行間で重ならない
      expect(versionA1).toBeLessThan(versionB1);
      expect(versionB1).toBeLessThan(versionA2);
    });

    it("assigns_versions_in_seed_order_when_rows_are_seeded", async () => {
      // Given: ep-a、ep-b の順で seed した repository
      const repository = new InMemoryProgressRepository([rowA, rowB]);

      // When: 各行の版を読む
      const versionA = await readVersion(repository, "ep-a");
      const versionB = await readVersion(repository, "ep-b");

      // Then: seed の並び順で版が増える
      expect(versionA).toBeLessThan(versionB);
    });

    it("advances_version_when_legacy_upsert_progress_replaces_row", async () => {
      // Given: ep-a を seed し、その時点の版を読んだ repository
      const repository = new InMemoryProgressRepository([rowA]);
      const versionBefore = await readVersion(repository, "ep-a");

      // When: 旧面の upsertProgress で同じ行を置き換える
      await repository.upsertProgress(replacedA);
      const versionAfter = await readVersion(repository, "ep-a");
      const got = await repository.replaceProgressIfVersion({
        row: rowA,
        expectedVersion: versionBefore,
      });

      // Then: 版が進み、書込前の版を期待値にした条件付き書込は conflict になる
      expect(versionAfter).toBeGreaterThan(versionBefore);
      expect(got).toBe("conflict");
    });

    it("keeps_version_independent_per_instance_when_two_repositories_are_seeded_alike", async () => {
      // Given: 同じ seed の repository を 2 つ作り、片方だけ書込を重ねる
      const first = new InMemoryProgressRepository([rowA]);
      const second = new InMemoryProgressRepository([rowA]);
      await first.upsertProgress(replacedA);

      // When: 書込を重ねていない側の版を読む
      const versionOfSecond = await readVersion(second, "ep-a");
      const versionOfFresh = await readVersion(new InMemoryProgressRepository([rowA]), "ep-a");

      // Then: 別 instance の書込に影響されず、新規 instance と同じ版になる
      expect(versionOfSecond).toBe(versionOfFresh);
    });
  });

  describe("listChangedAfter", () => {
    it("returns_rows_in_write_order_with_last_row_version_as_cursor_when_called_from_origin", async () => {
      // Given: ep-a、ep-b の順で書き、その後 ep-a を書き直した repository
      const repository = new InMemoryProgressRepository();
      await repository.insertProgressIfAbsent({ row: rowA });
      const versionA1 = await readVersion(repository, "ep-a");
      await repository.insertProgressIfAbsent({ row: rowB });
      await repository.replaceProgressIfVersion({ row: replacedA, expectedVersion: versionA1 });
      const versionA2 = await readVersion(repository, "ep-a");

      // When: 起点の cursor から差分取得する
      const got = await repository.listChangedAfter({ cursor: originCursor });

      // Then: 書き直した ep-a は最後に来て、cursor は最後の行の版になる
      expect(got.entries.map((entry) => entry.episodeId)).toEqual(["ep-b", "ep-a"]);
      expect(got.entries[1]?.progress).toEqual({
        positionSec: 99,
        firstPlayedAt: "2026-09-15T00:00:00.000Z",
        firstCompletedAt: "2026-09-16T00:00:00.000Z",
        lastPlayedAt: "2026-09-17T00:00:00.000Z",
      });
      expect(got.cursor).toBe(versionA2);
    });

    it("returns_seeded_rows_in_seed_order_when_called_from_origin", async () => {
      // Given: ep-a、ep-b の順で seed した repository
      const repository = new InMemoryProgressRepository([rowA, rowB]);

      // When: 起点の cursor から差分取得する
      const got = await repository.listChangedAfter({ cursor: originCursor });

      // Then: seed の並び順で返る
      expect(got.entries.map((entry) => entry.episodeId)).toEqual(["ep-a", "ep-b"]);
    });

    it("returns_only_rows_written_after_cursor_when_cursor_is_returned_by_previous_call", async () => {
      // Given: 1 回差分取得した後に、別の episode を書いた repository
      const repository = new InMemoryProgressRepository([rowA]);
      const previous = await repository.listChangedAfter({ cursor: originCursor });
      await repository.insertProgressIfAbsent({ row: rowB });

      // When: 前回の cursor から差分取得する
      const got = await repository.listChangedAfter({ cursor: previous.cursor });

      // Then: 後から書いた行だけが返る
      expect(got.entries.map((entry) => entry.episodeId)).toEqual(["ep-b"]);
    });

    it("returns_empty_entries_and_same_cursor_when_nothing_changed_after_cursor", async () => {
      // Given: 差分取得して最新の cursor を得た repository
      const repository = new InMemoryProgressRepository([rowA, rowB]);
      const previous = await repository.listChangedAfter({ cursor: originCursor });

      // When: その cursor で再び差分取得する
      const got = await repository.listChangedAfter({ cursor: previous.cursor });

      // Then: 空の entries と、渡した cursor がそのまま返る
      expect(got).toEqual({ cursor: previous.cursor, entries: [] });
    });

    it("returns_origin_cursor_and_empty_entries_when_repository_is_empty", async () => {
      // Given: 空の repository
      const repository = new InMemoryProgressRepository();

      // When: 起点の cursor から差分取得する
      const got = await repository.listChangedAfter({ cursor: originCursor });

      // Then: 空の entries と、渡した起点の cursor がそのまま返る
      expect(got).toEqual({ cursor: originCursor, entries: [] });
    });
  });
});
