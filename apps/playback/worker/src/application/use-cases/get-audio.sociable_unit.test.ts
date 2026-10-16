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
    getManuscript: async () => {
      throw new Error("not used");
    },
    getAudio: async () => validAudioBytes,
    ...overrides,
  };
}

describe("getAudio", () => {
  it("Port が返した mp3 バイト列をそのまま返す", async () => {
    // Given: 音声 byte を返す Fake Port
    const repository = createFakeRepository();

    // When: 音声取得 UseCase を実行する
    const got = await getAudio(repository, { episodeId: "ep-1" });

    // Then: byte が一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("Input の episodeId で Port に音声を問い合わせる", async () => {
    // Given: 受け取った引数を記録する Fake Port
    const received: unknown[] = [];
    const repository = createFakeRepository({
      getAudio: async (args) => {
        received.push(args);
        return validAudioBytes;
      },
    });

    // When: 音声取得 UseCase を実行する
    await getAudio(repository, { episodeId: "ep-7" });

    // Then: Port へは episodeId だけの引数 object を 1 回渡す
    expect(received).toEqual([{ episodeId: "ep-7" }]);
  });

  it("Port が undefined（mp3 無し）を返す時、EpisodeContentError（音声が無い）", async () => {
    // Given: mp3 無し
    const repository = createFakeRepository({ getAudio: async () => undefined });

    // When: 音声取得 UseCase を実行する
    const act = getAudio(repository, { episodeId: "ep-1" });

    // Then: Domain の実体不備
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/音声が無い/);
  });

  it("Port が mp3 でない byte を返す時、EpisodeContentError（音声が mp3 でない）", async () => {
    // Given: 署名の無い byte を返す Fake Port
    const repository = createFakeRepository({
      getAudio: async () => new Uint8Array([0x00, 0x01, 0x02]),
    });

    // When: 音声取得 UseCase を実行する
    const act = getAudio(repository, { episodeId: "ep-1" });

    // Then: 内容不適合として Domain の実体不備になる
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/音声が mp3 でない/);
  });

  it("Port が空の byte を返す時、EpisodeContentError（音声が空）", async () => {
    // Given: 0 byte を返す Fake Port
    const repository = createFakeRepository({ getAudio: async () => new Uint8Array([]) });

    // When: 音声取得 UseCase を実行する
    const act = getAudio(repository, { episodeId: "ep-1" });

    // Then: 不在（undefined）とは別の message で内容不適合になる
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
    await expect(act).rejects.toThrow(/音声が空/);
  });
});
