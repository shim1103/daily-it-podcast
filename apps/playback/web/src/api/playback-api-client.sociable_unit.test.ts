import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../../../contracts/index.ts";
import type { ApiResult } from "./api-result.ts";
import { buildRequestUrl, createPlaybackApiClient } from "./playback-api-client.ts";
import { requestBlob, requestJson } from "./playback-api-response.ts";

vi.mock("./playback-api-response.ts", () => ({
  requestBlob: vi.fn(),
  requestJson: vi.fn(),
}));

const requestJsonStub = vi.mocked(requestJson);
const requestBlobStub = vi.mocked(requestBlob);

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
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("listEpisodes は一覧 URL と一覧 schema を response module へ渡し Result を返す", async () => {
    // Given: response module が返す一覧 Result
    const fetch = vi.fn();
    const result: ApiResult<{ episodes: never[] }> = { ok: true, data: { episodes: [] } };
    requestJsonStub.mockResolvedValue(result);
    const client = createPlaybackApiClient({
      baseUrl: "https://example.test/",
      fetch,
    });

    // When: 一覧 endpoint を呼ぶ
    const got = await client.listEpisodes();

    // Then: URL・schema を委譲し、Result をそのまま返す
    expect(requestJsonStub).toHaveBeenCalledWith(
      fetch,
      "https://example.test/episodes",
      ListEpisodesResponseSchema,
    );
    expect(got).toBe(result);
  });

  it("getEpisode は episode URL と 1 件 schema を response module へ渡し Result を返す", async () => {
    // Given: response module が返す endpoint Result
    const fetch = vi.fn();
    const result: ApiResult<never> = { ok: false, error: "episode_not_found" };
    requestJsonStub.mockResolvedValue(result);
    const client = createPlaybackApiClient({
      baseUrl: "https://example.test",
      fetch,
    });

    // When: 1 件 endpoint を呼ぶ
    const got = await client.getEpisode("ep 1");

    // Then: URL・schema を委譲し、Result をそのまま返す
    expect(requestJsonStub).toHaveBeenCalledWith(
      fetch,
      "https://example.test/episodes/ep%201",
      GetEpisodeResponseSchema,
    );
    expect(got).toBe(result);
  });

  it("fetchAudio は audioRef URL を Blob response module へ渡し Result を返す", async () => {
    // Given: response module が返す音声 Result
    const fetch = vi.fn();
    const result: ApiResult<Blob> = { ok: false, error: "invalid_response" };
    requestBlobStub.mockResolvedValue(result);
    const client = createPlaybackApiClient({
      baseUrl: "https://example.test/",
      fetch,
    });

    // When: 音声 endpoint を呼ぶ
    const got = await client.fetchAudio("/episodes/ep-1/audio");

    // Then: audioRef URL を委譲し、Result をそのまま返す
    expect(requestBlobStub).toHaveBeenCalledWith(fetch, "https://example.test/episodes/ep-1/audio");
    expect(got).toBe(result);
  });
});
