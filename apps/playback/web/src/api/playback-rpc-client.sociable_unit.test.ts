import { describe, expect, it } from "vitest";
import {
  episodeProgressCompletePath,
  episodeProgressPath,
  listEpisodesPath,
  progressPullPath,
} from "../../../contracts/index.ts";
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

  it("createProgress は POST で episodeProgressPath を fetch する", async () => {
    // Given: method / URL を記録する fetch Stub
    let requestedUrl: string | undefined;
    let requestedMethod: string | undefined;
    const fetch = (url: string, init?: RequestInit) => {
      requestedUrl = url;
      requestedMethod = init?.method;
      return Promise.resolve(new Response(null, { status: 200 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test/", fetch });

    // When: create を発行する
    await rpc.createProgress("ep-1", {
      positionSec: 1,
      clientAt: "2026-09-22T10:00:00.000Z",
    });

    // Then: 契約 path・POST
    expect(requestedUrl).toBe(`https://example.test${episodeProgressPath("ep-1")}`);
    expect(requestedMethod).toBe("POST");
  });

  it("updateProgress は PATCH で episodeProgressPath を fetch する", async () => {
    // Given: method を記録する fetch Stub
    let requestedMethod: string | undefined;
    const fetch = (_url: string, init?: RequestInit) => {
      requestedMethod = init?.method;
      return Promise.resolve(new Response(null, { status: 200 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test/", fetch });

    // When: update を発行する
    await rpc.updateProgress("ep-1", {
      positionSec: 2,
      clientAt: "2026-09-22T10:00:00.000Z",
    });

    // Then: PATCH
    expect(requestedMethod).toBe("PATCH");
  });

  it("completeProgress は POST で episodeProgressCompletePath を fetch する", async () => {
    // Given: URL を記録する fetch Stub
    let requestedUrl: string | undefined;
    const fetch = (url: string) => {
      requestedUrl = url;
      return Promise.resolve(new Response(null, { status: 200 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test/", fetch });

    // When: complete を発行する
    await rpc.completeProgress("ep-1", {
      positionSec: 57,
      clientAt: "2026-09-22T10:00:00.000Z",
    });

    // Then: complete path
    expect(requestedUrl).toBe(`https://example.test${episodeProgressCompletePath("ep-1")}`);
  });

  it("pullProgress は GET で progressPullPath と since query を fetch する", async () => {
    // Given: URL を記録する fetch Stub
    let requestedUrl: string | undefined;
    const fetch = (url: string) => {
      requestedUrl = url;
      return Promise.resolve(new Response(null, { status: 200 }));
    };
    const rpc = createPlaybackRpcClient({ baseUrl: "https://example.test/", fetch });
    const since = "2026-09-22T10:00:00.000Z";

    // When: pull を発行する
    await rpc.pullProgress({ since });

    // Then: pull path + since
    expect(requestedUrl).toBe(
      `https://example.test${progressPullPath}?since=${encodeURIComponent(since)}`,
    );
  });
});
