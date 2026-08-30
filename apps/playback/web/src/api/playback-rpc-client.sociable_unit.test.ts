import { describe, expect, it } from "vitest";
import { listEpisodesPath } from "../../../contracts/index.ts";
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
});
