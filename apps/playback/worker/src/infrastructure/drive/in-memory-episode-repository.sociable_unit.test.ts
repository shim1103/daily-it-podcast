import { describe, expect, it } from "vitest";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import { InMemoryEpisodeRepository } from "./in-memory-episode-repository.ts";

/**
 * scope: Sociable Unit
 * real: InMemoryEpisodeRepository
 * double: none
 */
const manuscriptJson = { episodeId: "ep-1", title: "題" };

describe("InMemoryEpisodeRepository", () => {
  it("put した json は listManuscripts で stem 付きの生 payload として取り出せる", async () => {
    // Given: json を格納（schema 適合かどうかは問わない）
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson, validAudioBytes);

    // When: 一覧を取得する
    const got = await repository.listManuscripts();

    // Then: 検証せず生のまま返す
    expect(got).toEqual([{ stem: "ep-1", json: manuscriptJson }]);
  });

  it("格納がゼロ件の時、listManuscripts は空配列を返す（null でない）", async () => {
    // Given: 空の repository
    const repository = new InMemoryEpisodeRepository();

    // When / Then
    const got = await repository.listManuscripts();
    expect(got).toEqual([]);
  });

  it("put した wav は getAudio でそのまま取り出せる", async () => {
    // Given: json + wav
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson, validAudioBytes);

    // When: 音声を取得する
    const got = await repository.getAudio("ep-1");

    // Then: 格納した byte と一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("wav 無し / 未格納の getAudio は undefined", async () => {
    // Given: json のみ / 空
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson);

    // When / Then: どちらも undefined
    expect(await repository.getAudio("ep-1")).toBeUndefined();
    expect(await repository.getAudio("missing")).toBeUndefined();
  });
});
