import { describe, expect, it } from "vitest";
import { StubProgressRepository } from "./progress-repository.ts";

/**
 * real: StubProgressRepository（in-memory／unit 用。永続せず zero value のみ返す）
 */
describe("StubProgressRepository", () => {
  it("getByEpisodeIds_always_returns_empty_map", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.getByEpisodeIds(["ep-1", "ep-2"]);

    expect(got.size).toBe(0);
  });

  it("writeProgress_returns_zero_value_first_timestamps", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.writeProgress({
      episodeId: "ep-1",
      positionSec: 12,
      clientAt: "2026-09-19T10:00:00.000Z",
    });

    expect(got).toEqual({ firstPlayedAt: "1970-01-01T00:00:00.000Z", firstCompletedAt: null });
  });

  it("completeProgress_returns_zero_value_first_timestamps", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.completeProgress({
      episodeId: "ep-1",
      positionSec: 60,
      clientAt: "2026-09-19T10:00:00.000Z",
    });

    expect(got).toEqual({ firstPlayedAt: "1970-01-01T00:00:00.000Z", firstCompletedAt: null });
  });

  it("listUpdatedSince_always_returns_empty_array", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.listUpdatedSince("2026-09-19T10:00:00.000Z");

    expect(got).toEqual([]);
  });
});
