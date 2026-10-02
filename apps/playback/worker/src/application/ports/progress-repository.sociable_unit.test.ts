import { describe, expect, it } from "vitest";
import { StubProgressRepository } from "./progress-repository.ts";

/**
 * scope: Sociable Unit
 * real: StubProgressRepository（in-memory／unit 用。何も永続しない）
 * double: なし
 */
describe("StubProgressRepository", () => {
  it("getByEpisodeIds_always_returns_empty_map", async () => {
    // Given: 何も永続しない Stub
    const stub = new StubProgressRepository();

    // When: 任意の id で取得する
    const got = await stub.getByEpisodeIds(["ep-1", "ep-2"]);

    // Then: 常に空
    expect(got.size).toBe(0);
  });

  it("upsertProgress_completes_without_persisting", async () => {
    // Given: 何も永続しない Stub
    const stub = new StubProgressRepository();

    // When: upsert した後に、同じ id を取得する
    const result = stub.upsertProgress({
      episodeId: "ep-1",
      positionSec: 12,
      firstPlayedAt: "2026-09-19T10:00:00.000Z",
      firstCompletedAt: null,
      lastPlayedAt: "2026-09-19T10:00:00.000Z",
    });

    // Then: reject せず完了し、永続されないので取得しても空
    await expect(result).resolves.toBeUndefined();
    expect((await stub.getByEpisodeIds(["ep-1"])).size).toBe(0);
  });

  it("listUpdatedSince_always_returns_empty_array", async () => {
    // Given: 何も永続しない Stub
    const stub = new StubProgressRepository();

    // When: 差分取得する
    const got = await stub.listUpdatedSince("2026-09-19T10:00:00.000Z");

    // Then: 常に空
    expect(got).toEqual([]);
  });
});
