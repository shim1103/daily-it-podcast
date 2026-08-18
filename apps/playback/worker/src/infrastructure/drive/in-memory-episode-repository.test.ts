import { describe, expect, it } from "vitest";
import {
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  episodeAudioPath,
} from "../../../../contracts/index.ts";
import { EpisodeNotFoundError } from "../../entities/errors/episode-not-found-error.ts";
import { getEpisode } from "../../application/use-cases/get-episode.ts";
import { getEpisodeAudio } from "../../application/use-cases/get-episode-audio.ts";
import { listEpisodes } from "../../application/use-cases/list-episodes.ts";
import { InMemoryEpisodeRepository } from "./in-memory-episode-repository.ts";
import { ManuscriptSchema } from "./manuscript-schema.ts";

const validManuscript = {
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

const audioBytes = new Uint8Array([0xff, 0xfb, 0x90]);

describe("InMemoryEpisodeRepository", () => {
  it("Get 成功時、返却原稿が manuscript schema に適合する", async () => {
    // Given: json + mp3 のペア
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript, audioBytes);

    // When: 1件取得する
    const got = await getEpisode(repository, "ep-1");

    // Then: schema 適合 + audioRef
    expect(GetEpisodeResponseSchema.safeParse(got).success).toBe(true);
    const { audioRef: _audioRef, ...manuscript } = got;
    expect(ManuscriptSchema.safeParse(manuscript).success).toBe(true);
    expect(got.audioRef).toBe(episodeAudioPath("ep-1"));
  });

  it("Get 音声成功時、mp3 byte が取得できる", async () => {
    // Given: json + mp3 のペア
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript, audioBytes);

    // When: 音声を取得する
    const got = await getEpisodeAudio(repository, "ep-1");

    // Then: byte が一致する
    expect(got).toEqual(audioBytes);
  });

  it("schema 不適合 JSON は List 行に含めない", async () => {
    // Given: 有効 JSON と不適合 JSON
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript, audioBytes);
    repository.put("bad", { episodeId: "bad" });

    // When: 一覧を取得する
    const got = await listEpisodes(repository);

    // Then: 有効分のみ
    expect(ListEpisodesResponseSchema.safeParse(got).success).toBe(true);
    expect(got.episodes).toHaveLength(1);
    expect(got.episodes[0]?.episodeId).toBe("ep-1");
  });

  it("stem と JSON 内 episodeId が不一致の件は List 行に含めない", async () => {
    // Given: stem と episodeId がズレた有効 JSON
    const repository = new InMemoryEpisodeRepository();
    repository.put("stem-a", { ...validManuscript, episodeId: "ep-other" }, audioBytes);

    // When: 一覧を取得する
    const got = await listEpisodes(repository);

    // Then: 行に出ない
    expect(got.episodes).toHaveLength(0);
  });

  it("存在しない episodeId の Get は EpisodeNotFoundError になる", async () => {
    // Given: 空の repository
    const repository = new InMemoryEpisodeRepository();

    // When: 存在しない id を Get する
    const act = getEpisode(repository, "missing");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("schema 不適合 JSON の Get は EpisodeNotFoundError になる", async () => {
    // Given: 音声はあるが JSON が schema 不適合
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", { episodeId: "ep-1" }, audioBytes);

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("mp3 が無い json のみ件は Get で EpisodeNotFoundError になる", async () => {
    // Given: 音声無しの有効 JSON
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript);

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("stem と JSON 内 episodeId が不一致の Get は EpisodeNotFoundError になる", async () => {
    // Given: stem と episodeId がズレた json + mp3
    const repository = new InMemoryEpisodeRepository();
    repository.put("stem-a", { ...validManuscript, episodeId: "ep-other" }, audioBytes);

    // When: stem で 1件取得する
    const act = getEpisode(repository, "stem-a");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("存在しない episodeId の Get 音声は EpisodeNotFoundError になる", async () => {
    // Given: 空の repository
    const repository = new InMemoryEpisodeRepository();

    // When: 存在しない id の音声を取得する
    const act = getEpisodeAudio(repository, "missing");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("エントリはあるが audio が無い Get 音声は EpisodeNotFoundError になる", async () => {
    // Given: json のみ（audio 無し）
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript);

    // When: 音声を取得する
    const act = getEpisodeAudio(repository, "ep-1");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });
});
