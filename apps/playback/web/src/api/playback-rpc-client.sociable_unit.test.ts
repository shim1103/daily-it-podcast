import { describe, expect, it } from "vitest";
import { episodePath, listEpisodesPath } from "../../../contracts/index.ts";
import { createPlaybackRpcClient } from "./playback-rpc-client.ts";

describe("createPlaybackRpcClient", () => {
  it("listEpisodes は listEpisodesPath と同じ wire path を fetch する", async () => {
    // Given: URL を記録する fetch Stub
    let requestedUrl: string | undefined;
    const fetch = (url: string) => {
      requestedUrl = url;
      return Promise.resolve(new Response(null, { status: 200 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test/", fetch });

    // When: 一覧 request を発行する
    await rpc.listEpisodes();

    // Then: 契約 listEpisodesPath と同じ URL
    expect(requestedUrl).toBe(`https://example.test${listEpisodesPath}`);
  });

  it("getEpisode は契約 episodePath（encodeURIComponent）と同じ wire path を fetch する", async () => {
    // Given: episode URL への fetch Stub
    let requestedUrl: string | undefined;
    const fetch = (url: string) => {
      requestedUrl = url;
      return Promise.resolve(new Response(null, { status: 404 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test", fetch });

    // When: スペースを含む episodeId で 1 件 request を発行する
    await rpc.getEpisode("ep 1");

    // Then: 契約 episodePath と同じ encode 済み URL
    expect(requestedUrl).toBe(`https://example.test${episodePath("ep 1")}`);
  });

  it("getEpisode は非 ASCII と path 区切りを含む episodeId でも episodePath と一致する URL を fetch する", async () => {
    // Given: 複数の特殊文字を含む episodeId と URL を記録する Stub
    const requestedUrls: string[] = [];
    const fetch = (url: string) => {
      requestedUrls.push(url);
      return Promise.resolve(new Response(null, { status: 404 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test", fetch });

    // When: 非 ASCII・`/` を含む ID で getEpisode する
    await rpc.getEpisode("エピソード1");
    await rpc.getEpisode("a/b");

    // Then: いずれも契約 episodePath と同じ wire path
    expect(requestedUrls).toEqual([
      `https://example.test${episodePath("エピソード1")}`,
      `https://example.test${episodePath("a/b")}`,
    ]);
  });
});
