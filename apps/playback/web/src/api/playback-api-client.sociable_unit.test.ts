import { describe, expect, it } from "vitest";
import { listEpisodesPath } from "../../../contracts/index.ts";
import { buildRequestUrl, createPlaybackApiClient } from "./playback-api-client.ts";

function stubFetch(): (input: string, init?: RequestInit) => Promise<Response> {
  return () => Promise.resolve(new Response(null, { status: 200 }));
}

describe("buildRequestUrl", () => {
  it("baseUrl の末尾に / が無い時、そのまま path を続ける", () => {
    // Given: 末尾 / の無い baseUrl
    const baseUrl = "https://example.test/api";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 段の区切りは 1 つ
    expect(got).toBe("https://example.test/api/episodes");
  });

  it("baseUrl の末尾に / が 1 つある時、重ねずに繋ぐ", () => {
    // Given: 末尾 / が 1 つの baseUrl
    const baseUrl = "https://example.test/api/";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 段の区切りは 1 つ
    expect(got).toBe("https://example.test/api/episodes");
  });

  it("baseUrl の末尾に / が複数ある時、全て落として繋ぐ", () => {
    // Given: 末尾 / が複数の baseUrl
    const baseUrl = "https://example.test/api///";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 段の区切りは 1 つ
    expect(got).toBe("https://example.test/api/episodes");
  });

  it("baseUrl が空文字の時、path だけを返す", () => {
    // Given: 同一 origin を指す空の baseUrl
    const baseUrl = "";

    // When: 契約 path を繋ぐ
    const got = buildRequestUrl(baseUrl, listEpisodesPath);

    // Then: 契約 path がそのまま残る
    expect(got).toBe("/episodes");
  });
});

describe("createPlaybackApiClient", () => {
  it("deps を渡した時、3 つの method を持つ client を返す", () => {
    // Given: baseUrl と注入する fetch
    const deps = { baseUrl: "https://example.test", fetch: stubFetch() };

    // When: client を組み立てる
    const client = createPlaybackApiClient(deps);

    // Then: 契約の 3 method が生えている
    expect(typeof client.listEpisodes).toBe("function");
    expect(typeof client.getEpisode).toBe("function");
    expect(typeof client.fetchAudio).toBe("function");
  });

  it("listEpisodes を呼んだ時、baseUrl と一覧 path を繋いだ URL を fetch へ渡す", async () => {
    // Given: 呼ばれた URL を記録する fetch
    const calledUrls: string[] = [];
    const client = createPlaybackApiClient({
      baseUrl: "https://example.test/",
      fetch: (input) => {
        calledUrls.push(input);
        return Promise.resolve(new Response(null, { status: 200 }));
      },
    });

    // When: 一覧を取る
    await client.listEpisodes();

    // Then: 契約 path が 1 度だけ叩かれる
    expect(calledUrls).toEqual(["https://example.test/episodes"]);
  });

  it("getEpisode を呼んだ時、episodeId を載せた URL を fetch へ渡す", async () => {
    // Given: 呼ばれた URL を記録する fetch
    const calledUrls: string[] = [];
    const client = createPlaybackApiClient({
      baseUrl: "https://example.test",
      fetch: (input) => {
        calledUrls.push(input);
        return Promise.resolve(new Response(null, { status: 200 }));
      },
    });

    // When: 1 件を取る
    await client.getEpisode("ep-1");

    // Then: 契約 path が 1 度だけ叩かれる
    expect(calledUrls).toEqual(["https://example.test/episodes/ep-1"]);
  });

  it("fetchAudio を呼んだ時、受け取った audioRef を baseUrl へ繋いで fetch へ渡す", async () => {
    // Given: 呼ばれた URL を記録する fetch
    const calledUrls: string[] = [];
    const client = createPlaybackApiClient({
      baseUrl: "https://example.test",
      fetch: (input) => {
        calledUrls.push(input);
        return Promise.resolve(new Response(null, { status: 200 }));
      },
    });

    // When: 音声を取る
    await client.fetchAudio("/episodes/ep-1/audio");

    // Then: audioRef をそのまま繋いだ URL が 1 度だけ叩かれる
    expect(calledUrls).toEqual(["https://example.test/episodes/ep-1/audio"]);
  });
});
