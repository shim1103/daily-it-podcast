import { describe, expect, it } from "vitest";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import { getAudio } from "./get-audio.ts";

/**
 * scope: Sociable Unit
 * real: getAudio use-case
 * double: EpisodeRepository を Fake Port に差し替え
 */
function createFakeRepository(overrides: Partial<EpisodeRepository> = {}): EpisodeRepository {
  return {
    listManuscripts: async () => {
      throw new Error("not used");
    },
    getAudio: async () => validAudioBytes,
    ...overrides,
  };
}

describe("getAudio", () => {
  it("Port が返した wav バイト列をそのまま返す", async () => {
    // Given: 音声 byte を返す Fake Port
    const repository = createFakeRepository();

    // When: 音声取得 UseCase を実行する
    const got = await getAudio(repository, "ep-1");

    // Then: byte が一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("Port が undefined（wav 無し）を返す時、EpisodeContentError（音声が無い）", async () => {
    // Given: wav 無し
    const repository = createFakeRepository({ getAudio: async () => undefined });

    // When: 音声取得 UseCase を実行する
    const act = getAudio(repository, "ep-1");

    // Then: Domain の実体不備
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/音声が無い/);
  });
});
