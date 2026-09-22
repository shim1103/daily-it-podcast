import { describe, expect, it } from "vitest";
import { StubProgressRepository } from "./progress-repository.ts";

/**
 * real: StubProgressRepository（A 足場。zero value のみ）
 */
describe("StubProgressRepository", () => {
  it("getByEpisodeIds は常に空 Map を返す", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.getByEpisodeIds(["ep-1", "ep-2"]);

    expect(got.size).toBe(0);
  });

  it("writeProgress は first* の zero value を返す", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.writeProgress({
      episodeId: "ep-1",
      positionSec: 12,
      clientAt: "2026-09-19T10:00:00.000Z",
    });

    expect(got).toEqual({ firstPlayedAt: "", firstCompletedAt: null });
  });

  it("completeProgress は first* の zero value を返す", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.completeProgress({
      episodeId: "ep-1",
      positionSec: 60,
      clientAt: "2026-09-19T10:00:00.000Z",
    });

    expect(got).toEqual({ firstPlayedAt: "", firstCompletedAt: null });
  });

  it("listUpdatedSince は常に空配列を返す", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.listUpdatedSince("2026-09-19T10:00:00.000Z");

    expect(got).toEqual([]);
  });
});
