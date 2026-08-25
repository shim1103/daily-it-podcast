import { describe, expect, it } from "vitest";
import { createPlaybackApiClient } from "./playback-api-client.ts";

describe("createPlaybackApiClient", () => {
  it("listEpisodes は一覧 URL を fetch し、成功 response を schema 検証済み Result で返す", async () => {
    // Given: 一覧 URL への fetch が成功 response を返す Stub
    const validResponse = {
      episodes: [{ episodeId: "ep-1", date: "2026-08-20", title: "今日の IT", durationSec: 60 }],
    };
    let requestedUrl: string | undefined;
    const fetch = (url: string) => {
      requestedUrl = url;
      return Promise.resolve(Response.json(validResponse));
    };
    const client = createPlaybackApiClient({ baseUrl: "https://example.test/", fetch });

    // When: 一覧 endpoint を呼ぶ
    const got = await client.listEpisodes();

    // Then: 一覧 URL を fetch し、schema 検証済み data を返す
    expect(requestedUrl).toBe("https://example.test/episodes");
    expect(got).toEqual({ ok: true, data: validResponse });
  });

  it("getEpisode は episode URL を fetch し、非成功 response を契約 error の Result で返す", async () => {
    // Given: episode URL への fetch が 404 response を返す Stub
    let requestedUrl: string | undefined;
    const fetch = (url: string) => {
      requestedUrl = url;
      return Promise.resolve(new Response(null, { status: 404 }));
    };
    const client = createPlaybackApiClient({ baseUrl: "https://example.test", fetch });

    // When: 1 件 endpoint を呼ぶ
    const got = await client.getEpisode("ep 1");

    // Then: episodeId を URL encode した episode URL を fetch し、契約 error を返す
    expect(requestedUrl).toBe("https://example.test/episodes/ep%201");
    expect(got).toEqual({ ok: false, error: "episode_not_found" });
  });
});
