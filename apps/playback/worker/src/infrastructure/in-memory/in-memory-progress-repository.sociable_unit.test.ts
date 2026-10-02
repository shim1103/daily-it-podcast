import { describe, expect, it } from "vitest";
import type { ProgressUpsertRow } from "../../application/ports/progress-repository.ts";
import { InMemoryProgressRepository } from "./in-memory-progress-repository.ts";

/**
 * scope: Sociable Unit
 * real: InMemoryProgressRepository
 * double: なし
 */
const rowA: ProgressUpsertRow = {
  episodeId: "ep-a",
  positionSec: 10,
  firstPlayedAt: "2026-09-01T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-09-10T12:00:00.000Z",
};

const rowB: ProgressUpsertRow = {
  episodeId: "ep-b",
  positionSec: 20,
  firstPlayedAt: "2026-09-02T00:00:00.000Z",
  firstCompletedAt: "2026-09-05T00:00:00.000Z",
  lastPlayedAt: "2026-09-20T12:00:00.000Z",
};

describe("InMemoryProgressRepository", () => {
  it("getByEpisodeIds_returns_seeded_rows_and_omits_missing_ids", async () => {
    // Given: seed 済みの repository
    const repository = new InMemoryProgressRepository([rowA, rowB]);

    // When: 存在する id と存在しない id を混ぜて取得する
    const got = await repository.getByEpisodeIds(["ep-a", "ep-missing", "ep-b"]);

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
    const replaced: ProgressUpsertRow = {
      episodeId: "ep-a",
      positionSec: 99,
      firstPlayedAt: "2026-09-15T00:00:00.000Z",
      firstCompletedAt: "2026-09-16T00:00:00.000Z",
      lastPlayedAt: "2026-09-17T00:00:00.000Z",
    };
    await repository.upsertProgress(replaced);
    const got = await repository.getByEpisodeIds(["ep-a"]);

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

  // todo: C が実装（in-memory は単調増加する版の採番）へ置換したら、この足場 describe を behavior test へ置き換える
  describe("scaffold: versioned surface", () => {
    it("get_progress_with_version_returns_null_even_when_row_is_seeded", async () => {
      // Given: ep-a を seed した repository
      const repository = new InMemoryProgressRepository([rowA]);

      // When: 版つきで取得する
      const got = await repository.getProgressWithVersion("ep-a");

      // Then: null を返す
      expect(got).toBeNull();
    });

    it("upsert_progress_if_version_returns_conflict_and_keeps_store_unchanged", async () => {
      // Given: 空の repository
      const repository = new InMemoryProgressRepository();

      // When: 行なしと見て期待値 null で条件付き書込する
      const got = await repository.upsertProgressIfVersion(rowA, null);
      const stored = await repository.getByEpisodeIds(["ep-a"]);

      // Then: "conflict" を返し、行は書かれない
      expect(got).toBe("conflict");
      expect(stored.size).toBe(0);
    });

    it("list_changed_after_returns_same_cursor_and_empty_entries_even_when_rows_are_seeded", async () => {
      // Given: 2 行を seed した repository
      const repository = new InMemoryProgressRepository([rowA, rowB]);

      // When: cursor を渡して差分取得する
      const got = await repository.listChangedAfter("0");

      // Then: 引数の cursor と空 entries を返す
      expect(got).toEqual({ cursor: "0", entries: [] });
    });
  });
});
