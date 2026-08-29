import { describe, expect, it } from "vitest";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import { getEpisodeAudio } from "./get-episode-audio.ts";

/**
 * scope: Sociable Unit
 * real: getEpisodeAudio use-case
 * double: EpisodeRepository を Fake Port に差し替え
 *
 * why: Port は wav byte か「無し（undefined）」を返すだけ。use-case は undefined を
 * EpisodeContentError へ写し、byte はそのまま返す。
 */
function createFakeRepository(overrides: Partial<EpisodeRepository> = {}): EpisodeRepository {
  return {
    listManuscripts: async () => {
      throw new Error("not used");
    },
    getManuscript: async () => {
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

  it("Port が undefined（wav 無し）を返す時、EpisodeContentError（音声が無い）", async () => {
    // Given: wav 無し
    const repository = createFakeRepository({ getEpisodeAudio: async () => undefined });

    // When: 音声取得 UseCase を実行する
    const act = getEpisodeAudio(repository, "ep-1");

    // Then: Domain の実体不備
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/音声が無い/);
  });
});
