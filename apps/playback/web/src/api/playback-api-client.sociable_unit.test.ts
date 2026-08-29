import { describe, expect, it } from "vitest";
import { createPlaybackApiClient } from "./playback-api-client.ts";

describe("createPlaybackApiClient", () => {
  it("listEpisodes は成功 response を schema 検証済み Result で返す", async () => {
    // Given: 一覧 URL への fetch が成功 response を返す Stub
    const validResponse = {
      episodes: [
        {
          episodeId: "ep-1",
          date: "2026-08-20",
          title: "今日の IT",
          durationSec: 60,
          topics: [{ title: "題" }],
          audioRef: "/episodes/ep-1/audio",
        },
      ],
    };
    const fetch = () => Promise.resolve(Response.json(validResponse));
    const client = createPlaybackApiClient({ baseUrl: "https://example.test/", fetch });

    // When: 一覧 endpoint を呼ぶ
    const got = await client.listEpisodes();

    // Then: schema 検証済み data を返す
    expect(got).toEqual({ ok: true, data: validResponse });
  });

  it("getEpisode は非成功 response を契約 error の Result で返す", async () => {
    // Given: episode URL への fetch が 404 response を返す Stub
    const fetch = () => Promise.resolve(new Response(null, { status: 404 }));
    const client = createPlaybackApiClient({ baseUrl: "https://example.test", fetch });

    // When: 1 件 endpoint を呼ぶ
    const got = await client.getEpisode("ep 1");

    // Then: 契約 error を返す
    expect(got).toEqual({ ok: false, error: "episode_not_found" });
  });

  it("getEpisode は成功 response を schema 検証済み Result で返す", async () => {
    // Given: 契約どおりの 1 件 body
    const validResponse = {
      episodeId: "ep-1",
      date: "2026-08-20",
      title: "今日の IT",
      durationSec: 60,
      body: {
        opening: "開始",
        topics: [{ title: "題", preface: "前", detail: "詳", startSec: 0 }],
        closing: "終了",
      },
      audioRef: "/episodes/ep-1/audio",
    };
    const fetch = () => Promise.resolve(Response.json(validResponse));
    const client = createPlaybackApiClient({ baseUrl: "https://example.test", fetch });

    // When: 1 件 endpoint を呼ぶ
    const got = await client.getEpisode("ep-1");

    // Then: schema 検証済み data を返す
    expect(got).toEqual({ ok: true, data: validResponse });
  });

  it("listEpisodes は network failure を network_error Result で返す", async () => {
    // Given: fetch が reject する Stub
    const fetch = () => Promise.reject(new TypeError("Failed to fetch"));
    const client = createPlaybackApiClient({ baseUrl: "https://example.test", fetch });

    // When: 一覧 endpoint を呼ぶ
    const got = await client.listEpisodes();

    // Then: throw せず network_error を返す
    expect(got).toEqual({ ok: false, error: "network_error" });
  });
});
