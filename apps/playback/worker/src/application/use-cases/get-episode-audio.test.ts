import { describe, expect, it } from "vitest";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import { getEpisodeAudio } from "./get-episode-audio.ts";

const audioBytes = new Uint8Array([0xff, 0xfb, 0x90]);

function createFakeRepository(
  overrides: Partial<EpisodeRepository> = {},
): EpisodeRepository {
  return {
    listEpisodes: async () => {
      throw new Error("not used");
    },
    getEpisode: async () => {
      throw new Error("not used");
    },
    getEpisodeAudio: async () => audioBytes,
    ...overrides,
  };
}

describe("getEpisodeAudio", () => {
  it("Port が返した mp3 バイト列をそのまま返す", async () => {
    // Given: 音声 byte を返す Fake Port
    const repository = createFakeRepository();

    // When: 音声取得 UseCase を実行する
    const got = await getEpisodeAudio(repository, "ep-1");

    // Then: byte が一致する
    expect(got).toEqual(audioBytes);
  });
});
