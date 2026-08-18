import { describe, expect, it } from "vitest";
import { ListEpisodesResponseSchema } from "../../../../contracts/index.ts";
import type { EpisodeListItem, EpisodeRepository } from "../ports/episode-repository.ts";
import { listEpisodes } from "./list-episodes.ts";

const validListItem: EpisodeListItem = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
};

function createFakeRepository(
  overrides: Partial<EpisodeRepository> = {},
): EpisodeRepository {
  return {
    listEpisodes: async () => [validListItem],
    getEpisode: async () => {
      throw new Error("not used");
    },
    getEpisodeAudio: async () => {
      throw new Error("not used");
    },
    ...overrides,
  };
}

describe("listEpisodes", () => {
  it("Port が返した目録を ListEpisodesResponse に包んで返す", async () => {
    // Given: 1件返す Fake Port
    const repository = createFakeRepository();

    // When: 一覧 UseCase を実行する
    const got = await listEpisodes(repository);

    // Then: 契約 schema を満たす
    expect(ListEpisodesResponseSchema.safeParse(got).success).toBe(true);
    expect(got.episodes).toEqual([validListItem]);
  });
});
