import { describe, expect, it } from "vitest";
import { R2EpisodeRepository } from "./r2-episode-repository.ts";

describe("R2EpisodeRepository stub", () => {
  it("listManuscripts は空配列、getAudio は undefined", async () => {
    // Given: binding 未使用の stub
    const repository = new R2EpisodeRepository({
      bucket: {
        get: async () => null,
        put: async () => undefined,
        list: async () => ({ objects: [] }),
      },
    });

    // When / Then: 零値
    expect(await repository.listManuscripts()).toEqual([]);
    expect(await repository.getAudio("ep-1")).toBeUndefined();
  });
});
