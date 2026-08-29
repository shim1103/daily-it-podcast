import { describe, expect, it } from "vitest";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import { InMemoryEpisodeRepository } from "./in-memory-episode-repository.ts";

/**
 * scope: Sociable Unit
 * real: InMemoryEpisodeRepository
 * double: none
 *
 * why: この Adapter 固有の責務は「Map への格納・取り出し」と「entry / audio 在否の戻り値表現」
 * だけ。schema 不適合・stem 不一致・4 失敗ケースの網羅は use-case test へ移した
 * （`get-episode.sociable_unit.test.ts` / `list-episodes.sociable_unit.test.ts`）。
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

  it("put した json+wav は getManuscript で生 payload と hasAudio: true を返す", async () => {
    // Given: json + wav
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson, validAudioBytes);

    // When: 1件取得する
    const got = await repository.getManuscript("ep-1");

    // Then: 検証せず、wav 有無だけ添えて返す
    expect(got).toEqual({ json: manuscriptJson, hasAudio: true });
  });

  it("wav 無しで put した entry の getManuscript は hasAudio: false", async () => {
    // Given: json のみ
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson);

    // When: 1件取得する
    const got = await repository.getManuscript("ep-1");

    // Then: throw せず hasAudio: false
    expect(got).toEqual({ json: manuscriptJson, hasAudio: false });
  });

  it("格納されていない episodeId の getManuscript は undefined（取得対象なし）", async () => {
    // Given: 空の repository
    const repository = new InMemoryEpisodeRepository();

    // When / Then: throw せず undefined
    expect(await repository.getManuscript("missing")).toBeUndefined();
  });

  it("put した wav は getEpisodeAudio でそのまま取り出せる", async () => {
    // Given: json + wav
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson, validAudioBytes);

    // When: 音声を取得する
    const got = await repository.getEpisodeAudio("ep-1");

    // Then: 格納した byte と一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("wav 無し / 未格納の getEpisodeAudio は undefined", async () => {
    // Given: json のみ / 空
    const repository = new InMemoryEpisodeRepository();
    repository.put("ep-1", manuscriptJson);

    // When / Then: どちらも undefined
    expect(await repository.getEpisodeAudio("ep-1")).toBeUndefined();
    expect(await repository.getEpisodeAudio("missing")).toBeUndefined();
  });
});
