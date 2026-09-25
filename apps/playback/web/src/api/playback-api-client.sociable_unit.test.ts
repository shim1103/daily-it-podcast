import { describe, expect, it } from "vitest";
import { createPlaybackApiClient } from "./playback-api-client.ts";

const writeBody = {
  positionSec: 12,
  clientAt: "2026-09-22T10:00:00.000Z",
} as const;

const writeResponse = {
  firstPlayedAt: "2026-09-22T10:00:00.000Z",
  firstCompletedAt: null,
} as const;

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
          body: {
            opening: { text: "開始", startSec: 0 },
            topics: [{ title: "題", preface: "前", detail: "詳", startSec: 0 }],
            ending: { text: "終了", startSec: 55 },
          },
          audioRef: "/episodes/ep-1/audio",
          progress: null,
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

  it("listEpisodes は network failure を network_error Result で返す", async () => {
    // Given: fetch が reject する Stub
    const fetch = () => Promise.reject(new TypeError("Failed to fetch"));
    const client = createPlaybackApiClient({ baseUrl: "https://example.test", fetch });

    // When: 一覧 endpoint を呼ぶ
    const got = await client.listEpisodes();

    // Then: throw せず network_error を返す
    expect(got).toEqual({ ok: false, error: "network_error" });
  });

  it("createProgress は成功 response を schema 検証済み Result で返す", async () => {
    // Given: Write 成功 Stub
    const fetch = () => Promise.resolve(Response.json(writeResponse));
    const client = createPlaybackApiClient({ baseUrl: "https://example.test/", fetch });

    // When: create
    const got = await client.createProgress("ep-1", writeBody);

    // Then: ok data
    expect(got).toEqual({ ok: true, data: writeResponse });
  });

  it("updateProgress / completeProgress / pullProgress の signature が ApiResult を返す", async () => {
    // Given: Write / pull 成功 Stub
    const fetch = (url: string) => {
      if (url.includes("since=") || (url.endsWith("/progress") && !url.includes("/episodes/"))) {
        return Promise.resolve(Response.json({ episodes: [] }));
      }
      return Promise.resolve(Response.json(writeResponse));
    };
    const client = createPlaybackApiClient({ baseUrl: "https://example.test/", fetch });

    // When: 残り 3 method
    const updated = await client.updateProgress("ep-1", writeBody);
    const completed = await client.completeProgress("ep-1", writeBody);
    const pulled = await client.pullProgress({ since: "2026-09-22T10:00:00.000Z" });

    // Then: いずれも ok（足場。retry 振る舞いは C）
    expect(updated.ok).toBe(true);
    expect(completed.ok).toBe(true);
    expect(pulled).toEqual({ ok: true, data: { episodes: [] } });
  });
});
