import { describe, expect, it } from "vitest";
import { GetEpisodeResponseSchema, episodeAudioPath } from "../../../../contracts/index.ts";
import type { EpisodeManuscript, EpisodeRepository } from "../ports/episode-repository.ts";
import { getEpisode } from "./get-episode.ts";

const validManuscript: EpisodeManuscript = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      {
        title: "題",
        preface: "前置き",
        detail: "詳細",
        startSec: 0,
      },
    ],
    closing: "終了",
  },
};

function createFakeRepository(overrides: Partial<EpisodeRepository> = {}): EpisodeRepository {
  return {
    listEpisodes: async () => {
      throw new Error("not used");
    },
    getEpisode: async () => validManuscript,
    getEpisodeAudio: async () => {
      throw new Error("not used");
    },
    ...overrides,
  };
}

describe("getEpisode", () => {
  it("原稿に audioRef を付与して GetEpisodeResponse を返す", async () => {
    // Given: 原稿を返す Fake Port
    const repository = createFakeRepository();

    // When: 1件取得 UseCase を実行する
    const got = await getEpisode(repository, "ep-1");

    // Then: audioRef が path 定数どおり
    expect(GetEpisodeResponseSchema.safeParse(got).success).toBe(true);
    expect(got.audioRef).toBe(episodeAudioPath("ep-1"));
    expect(got.body).toEqual(validManuscript.body);
  });
});
