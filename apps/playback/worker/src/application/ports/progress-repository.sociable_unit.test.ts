import { describe, expect, it } from "vitest";
import { StubProgressRepository } from "./progress-repository.ts";

/**
 * real: StubProgressRepository（A 足場。空 Map のみ）
 */
describe("StubProgressRepository", () => {
  it("getByEpisodeIds は常に空 Map を返す", async () => {
    const stub = new StubProgressRepository();

    const got = await stub.getByEpisodeIds(["ep-1", "ep-2"]);

    expect(got.size).toBe(0);
  });
});
