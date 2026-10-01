import { describe, expect, it } from "vitest";
import { StubProgressRepository } from "./progress-repository.ts";

/**
 * scope: Sociable Unit
 * real: StubProgressRepository（Composition 用 zero）
 * double: なし
 */
describe("StubProgressRepository", () => {
  it("getByEpisodeIds は常に空 Map を返す", async () => {
    // Given: zero Stub
    const stub = new StubProgressRepository();

    // When: 任意の id で取得する
    const got = await stub.getByEpisodeIds(["ep-1", "ep-2"]);

    // Then: 常に空
    expect(got.size).toBe(0);
  });

  it("upsertProgress は no-op で完了する", async () => {
    // Given: zero Stub
    const stub = new StubProgressRepository();

    // When: upsert する
    const result = stub.upsertProgress({
      episodeId: "ep-1",
      positionSec: 12,
      firstPlayedAt: "2026-09-19T10:00:00.000Z",
      firstCompletedAt: null,
      lastPlayedAt: "2026-09-19T10:00:00.000Z",
    });

    // Then: reject しない（副作用なし）
    await expect(result).resolves.toBeUndefined();
  });

  it("listUpdatedSince は常に空配列を返す", async () => {
    // Given: zero Stub
    const stub = new StubProgressRepository();

    // When: 差分取得する
    const got = await stub.listUpdatedSince("2026-09-19T10:00:00.000Z");

    // Then: 常に空
    expect(got).toEqual([]);
  });
});
