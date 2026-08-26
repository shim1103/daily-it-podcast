import { beforeEach, describe, expect, it, vi } from "vitest";
import { ListEpisodesResponseSchema } from "../../../contracts/index.ts";
import type { PlaybackApiErrorCode } from "./playback-api-error.ts";
import { mapHttpStatusToApiError } from "./playback-api-error.ts";
import { readJsonResult } from "./playback-api-response.ts";

vi.mock("./playback-api-error.ts", () => ({
  mapHttpStatusToApiError: vi.fn(),
}));

const mapHttpStatusToApiErrorStub = vi.mocked(mapHttpStatusToApiError);

const validListEpisodesResponse = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-20",
      title: "今日の IT",
      durationSec: 60,
    },
  ],
};

describe("readJsonResult 正常系", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("合法な JSON response の時、schema 検証済み data を返す", async () => {
    // Given: 契約に適合する JSON を持つ成功 response
    const response = Response.json(validListEpisodesResponse);

    // When: JSON response を処理する
    const got = await readJsonResult(() => Promise.resolve(response), ListEpisodesResponseSchema);

    // Then: schema 検証済みの data を返す
    expect(got).toEqual({ ok: true, data: validListEpisodesResponse });
  });
});

describe("readJsonResult 異常系", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("getResponse が reject した時、network_error を返す", async () => {
    // Given: network failure を起こす getResponse
    // When: JSON response を処理する
    const got = await readJsonResult(
      () => Promise.reject(new Error("network failure")),
      ListEpisodesResponseSchema,
    );

    // Then: network_error
    expect(got).toEqual({ ok: false, error: "network_error" });
  });

  it("成功 JSON の schema 検証が失敗した時、invalid_response を返す", async () => {
    // Given: schema に合わない JSON を持つ成功 response
    const response = Response.json({ episodes: [{ invalid: true }] });

    // When: JSON response を処理する
    const got = await readJsonResult(() => Promise.resolve(response), ListEpisodesResponseSchema);

    // Then: invalid_response
    expect(got).toEqual({ ok: false, error: "invalid_response" });
  });

  it("成功 response の JSON 読み取りが reject した時、invalid_response を返す", async () => {
    // Given: json 読み取りに失敗する成功 response
    const response = {
      ok: true,
      status: 200,
      json: () => Promise.reject(new Error("JSON read failure")),
    };

    // When: JSON response を処理する
    const got = await readJsonResult(() => Promise.resolve(response), ListEpisodesResponseSchema);

    // Then: invalid_response
    expect(got).toEqual({ ok: false, error: "invalid_response" });
  });

  it("non-ok response の時、stub の error を返し failed body を読まない", async () => {
    // Given: web error mapper の Stub と body を読めない非成功 response
    const mappedError: PlaybackApiErrorCode = "configuration_error";
    mapHttpStatusToApiErrorStub.mockReturnValue(mappedError);
    let bodyRead = false;
    const response = {
      ok: false,
      status: 599,
      json: async () => {
        bodyRead = true;
        throw new Error("失敗 body は読まない");
      },
    };

    // When: JSON response を処理する
    const got = await readJsonResult(() => Promise.resolve(response), ListEpisodesResponseSchema);

    // Then: mapper の値をそのまま返し body は読まない
    expect(got).toEqual({ ok: false, error: mappedError });
    expect(mapHttpStatusToApiErrorStub).toHaveBeenCalledWith(599);
    expect(bodyRead).toBe(false);
  });
});

describe("readJsonResult 境界系", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("Response.ok が true の時、edge 2xx status を成功として body を読む", async () => {
    // Given: status は 299 だが ok が true の response
    const response = {
      ok: true,
      status: 299,
      json: () => Promise.resolve(validListEpisodesResponse),
    };

    // When: JSON response を処理する
    const got = await readJsonResult(() => Promise.resolve(response), ListEpisodesResponseSchema);

    // Then: status の分類ではなく ok を根拠に成功する
    expect(got).toEqual({ ok: true, data: validListEpisodesResponse });
  });
});
