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
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";

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

describe("InMemoryEpisodeRepository", () => {
  it("Get 成功時、返却原稿が manuscript schema に適合する", async () => {
    // Given: json + wav のペア
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript, validAudioBytes);

    // When: 1件取得する
    const got = await getEpisode(repository, "ep-1");

    // Then: schema 適合 + audioRef
    expect(GetEpisodeResponseSchema.safeParse(got).success).toBe(true);
    const { audioRef: _audioRef, ...manuscript } = got;
    expect(ManuscriptSchema.safeParse(manuscript).success).toBe(true);
    expect(got.audioRef).toBe(episodeAudioPath("ep-1"));
  });

  it("Get 音声成功時、wav byte が取得できる", async () => {
    // Given: json + wav のペア
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript, validAudioBytes);

    // When: 音声を取得する
    const got = await getEpisodeAudio(repository, "ep-1");

    // Then: byte が一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("List の topics[].title 列が同一 episode の Get body.topics[].title 列と順序込みで一致する", async () => {
    // Given: 複数 topic を持つ原稿
    const multiTopicManuscript = {
      ...validManuscript,
      body: {
        ...validManuscript.body,
        topics: [
          { title: "第一トピック", preface: "前1", detail: "詳1", startSec: 0 },
          { title: "第二トピック", preface: "前2", detail: "詳2", startSec: 30 },
          { title: "第三トピック", preface: "前3", detail: "詳3", startSec: 60 },
        ],
      },
    };
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", multiTopicManuscript, validAudioBytes);

    // When: 一覧と 1件を両方取得する
    const list = await listEpisodes(repository);
    const detail = await getEpisode(repository, "ep-1");

    // Then: list 側の題名列が Get 側の題名列と順序込みで一致する
    expect(list.episodes).toHaveLength(1);
    const listTitles = list.episodes[0]?.topics.map((topic) => topic.title);
    const detailTitles = detail.body.topics.map((topic) => topic.title);
    expect(listTitles).toEqual(detailTitles);
    expect(listTitles).toEqual(["第一トピック", "第二トピック", "第三トピック"]);
  });

  it("schema 不適合 JSON は List 行に含めない", async () => {
    // Given: 有効 JSON と不適合 JSON
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript, validAudioBytes);
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
    repository.put("stem-a", { ...validManuscript, episodeId: "ep-other" }, validAudioBytes);

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
    repository.put("ep-1", { episodeId: "ep-1" }, validAudioBytes);

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("wav が無い json のみ件は Get で EpisodeNotFoundError になる", async () => {
    // Given: 音声無しの有効 JSON
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", validManuscript);

    // When: 1件取得する
    const act = getEpisode(repository, "ep-1");

    // Then: Domain 不在
    await expect(act).rejects.toBeInstanceOf(EpisodeNotFoundError);
  });

  it("stem と JSON 内 episodeId が不一致の Get は EpisodeNotFoundError になる", async () => {
    // Given: stem と episodeId がズレた json + wav
    const repository = new InMemoryEpisodeRepository();
    repository.put("stem-a", { ...validManuscript, episodeId: "ep-other" }, validAudioBytes);

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
