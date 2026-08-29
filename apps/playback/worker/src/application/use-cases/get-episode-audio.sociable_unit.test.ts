import { describe, expect, it } from "vitest";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import { getEpisodeAudio } from "./get-episode-audio.ts";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";

function createFakeRepository(overrides: Partial<EpisodeRepository> = {}): EpisodeRepository {
  return {
    listEpisodes: async () => {
      throw new Error("not used");
    },
    getEpisode: async () => {
      throw new Error("not used");
    },
    getEpisodeAudio: async () => validAudioBytes,
    ...overrides,
  };
}

describe("getEpisodeAudio", () => {
  it("Port が返した wav バイト列をそのまま返す", async () => {
    // Given: 音声 byte を返す Fake Port
    const repository = createFakeRepository();

    // When: 音声取得 UseCase を実行する
    const got = await getEpisodeAudio(repository, "ep-1");

    // Then: byte が一致する
    expect(got).toEqual(validAudioBytes);
  });
});
