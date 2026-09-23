import { afterAll, afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorResponseSchema,
  episodeAudioContentType,
  episodeAudioPath,
  ListEpisodesResponseSchema,
  listEpisodesPath,
  NotFoundError,
} from "../../../contracts/index.ts";

const listEpisodesController = vi.fn();
const getAudioController = vi.fn();

vi.mock("../composition/root.ts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../composition/root.ts")>();
  return {
    ...actual,
    createPlaybackControllers: vi.fn(() => ({
      listEpisodesController,
      getAudioController,
    })),
  };
});

import { createPlaybackControllers, PlaybackRuntimeConfigError } from "../composition/root.ts";
import { app, createApp, throwOnEpisodeIdValidationFailure } from "./app.ts";
import { validAudioBytes } from "../test/fixtures/audio-bytes.ts";
import { requestIdHeaderName } from "./request-context.ts";

const origin = "http://example.test";
const emptyEnv = {};

const validList = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題",
      durationSec: 60,
      body: {
        opening: { text: "開始", startSec: 0 },
        topics: [
          {
            title: "題",
            preface: "前置き",
            detail: "詳細",
            startSec: 0,
          },
        ],
        ending: { text: "終了", startSec: 55 },
      },
      audioRef: episodeAudioPath("ep-1"),
      progress: null,
    },
  ],
};

const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

afterEach(() => {
  vi.mocked(listEpisodesController).mockReset();
  vi.mocked(getAudioController).mockReset();
  vi.mocked(createPlaybackControllers).mockClear();
  errorSpy.mockClear();
  logSpy.mockClear();
});

afterAll(() => {
  errorSpy.mockRestore();
  logSpy.mockRestore();
});

describe("throwOnEpisodeIdValidationFailure", () => {
  it("zValidator の parse が成功した時、何も throw しない", () => {
    // Given: 成功結果
    // When / Then: throw しない
    expect(() => throwOnEpisodeIdValidationFailure({ success: true })).not.toThrow();
  });

  it("zValidator の parse が失敗した時、ValidationError を throw し zod error を cause へ残す", () => {
    // Given: 失敗結果
    const zodError = new Error("zod validation failed");

    // When / Then: ValidationError を throw する
    expect(() =>
      throwOnEpisodeIdValidationFailure({ success: false, error: zodError }),
    ).toThrowError(
      expect.objectContaining({
        name: "ValidationError",
        message: "入力が契約に不適合",
        cause: zodError,
      }),
    );
  });
});

describe("app", () => {
  it("Hono instance を export する", () => {
    expect(app).toBeDefined();
    expect(typeof app.fetch).toBe("function");
  });

  it("createApp が useCaseOverrides を渡す時、その override で Hono instance を組み立てる", async () => {
    // Given: dev-only の fake use case override
    vi.mocked(listEpisodesController).mockResolvedValue(validList);
    const overrides = {
      useCases: {
        listEpisodes: vi.fn(),
        getAudio: vi.fn(),
      },
    };
    const devApp = createApp(overrides);

    // When: 一覧 path へ GET する
    await devApp.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: createPlaybackControllers へ override と R2 mode 固定がそのまま渡る
    expect(createPlaybackControllers).toHaveBeenCalledWith(emptyEnv, { mode: "r2" }, overrides);
  });

  it("受け取った env をそのまま Composition Root へ渡して Controller を組み立てる", async () => {
    // Given: 任意の binding を模した値
    vi.mocked(listEpisodesController).mockResolvedValue(validList);
    const boundEnv = { EPISODES: { get: async () => null, list: async () => ({ objects: [] }) } };

    // When: 一覧 path へ GET する
    await app.request(`${origin}${listEpisodesPath}`, {}, boundEnv);

    // Then: 渡された env と R2 mode 固定で Composition Root に渡る
    expect(createPlaybackControllers).toHaveBeenCalledWith(boundEnv, { mode: "r2" }, undefined);
  });

  it("一覧 GET が成功する時、ListEpisodesResponse schema を満たす JSON を 200 で返す", async () => {
    // Given: Composition が契約どおりの一覧を返す
    vi.mocked(listEpisodesController).mockResolvedValue(validList);

    // When: 一覧 path へ GET する
    const got = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: 200・契約 schema・progress embed のため no-store（browser も edge も）
    expect(got.status).toBe(200);
    expect(got.headers.get("Cache-Control")).toBe("no-store");
    expect(got.headers.get("Cloudflare-CDN-Cache-Control")).toBe("no-store");
    const body: unknown = await got.json();
    expect(ListEpisodesResponseSchema.safeParse(body).success).toBe(true);
  });

  it("一覧 GET が成功する時、ETag を付与する", async () => {
    // Given: Composition が契約どおりの一覧を返す
    vi.mocked(listEpisodesController).mockResolvedValue(validList);

    // When: 一覧 path へ GET する
    const got = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: ETag header が付く
    expect(got.headers.get("ETag")).toBeTruthy();
  });

  it("一覧 GET に If-None-Match を一致させて送る時、304 を body なしで返す", async () => {
    // Given: 1 回目で得た ETag
    vi.mocked(listEpisodesController).mockResolvedValue(validList);
    const first = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);
    const etag = first.headers.get("ETag");

    // When: 同じ ETag を If-None-Match として送る
    const second = await app.request(
      `${origin}${listEpisodesPath}`,
      { headers: { "If-None-Match": etag ?? "" } },
      emptyEnv,
    );

    // Then: 304・body なし
    expect(second.status).toBe(304);
    expect(await second.text()).toBe("");
  });

  it("音声 GET が成功する時、契約の Content-Type で byte を返す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    const got = await app.request(`${origin}${episodeAudioPath("ep-1")}`, {}, emptyEnv);

    // Then: JSON ではなく契約 Content-Type の byte
    expect(got.status).toBe(200);
    expect(got.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await got.arrayBuffer());
    expect(bytes).toEqual(validAudioBytes);
  });

  it("音声 GET が Range header 付きの時、206 で該当範囲だけ返す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getAudioController).mockResolvedValue(validAudioBytes);

    // When: Range header 付きで音声 path へ GET する
    const got = await app.request(
      `${origin}${episodeAudioPath("ep-1")}`,
      { headers: { Range: `bytes=1-2` } },
      emptyEnv,
    );

    // Then: 206・Content-Range・該当 2 byte
    expect(got.status).toBe(206);
    expect(got.headers.get("Content-Range")).toBe(`bytes 1-2/${validAudioBytes.length}`);
    const bytes = new Uint8Array(await got.arrayBuffer());
    expect(bytes).toEqual(validAudioBytes.subarray(1, 3));
  });

  it("音声 GET の path param を zValidator で検証済みの episodeId として Controller に渡す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    await app.request(`${origin}${episodeAudioPath("ep-1")}`, {}, emptyEnv);

    // Then: zValidator（EpisodeIdRequestSchema）を経由した検証済み episodeId で渡る
    expect(getAudioController).toHaveBeenCalledWith("ep-1");
  });

  it("音声 GET の Controller が NotFoundError を throw する時、404 と episode_not_found を返す", async () => {
    // Given: Domain 不在を写した External Error
    vi.mocked(getAudioController).mockRejectedValue(new NotFoundError("エピソードが無い"));

    // When: 音声 path へ GET する
    const got = await app.request(`${origin}${episodeAudioPath("missing")}`, {}, emptyEnv);

    // Then: 404 と契約 code のみ
    expect(got.status).toBe(404);
    const body: unknown = await got.json();
    expect(body).toEqual({ code: "episode_not_found" });
  });

  it("method または path が契約に無い時、400 と validation_error を返す", async () => {
    // Given: 契約に無い path
    // When: GET する
    const got = await app.request(`${origin}/unknown`, {}, emptyEnv);

    // Then: 未一致を episode_not_found に畳まない
    expect(got.status).toBe(400);
    const body: unknown = await got.json();
    expect(body).toEqual({ code: "validation_error" });
  });

  it("POST の一覧 path は 400 と validation_error を返す", async () => {
    // Given: 契約 method ではない POST
    // When: 送る
    const got = await app.request(`${origin}${listEpisodesPath}`, { method: "POST" }, emptyEnv);

    // Then: 未一致は validation_error
    expect(got.status).toBe(400);
    const body: unknown = await got.json();
    expect(body).toEqual({ code: "validation_error" });
  });

  it("一覧 GET が成功する時、X-Content-Type-Options: nosniff を付与する", async () => {
    // Given: Composition が契約どおりの一覧を返す
    vi.mocked(listEpisodesController).mockResolvedValue(validList);

    // When: 一覧 path へ GET する
    const got = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: secureHeaders() が MIME sniffing 対策 header を付与する
    expect(got.headers.get("X-Content-Type-Options")).toBe("nosniff");
  });

  it("音声 GET が成功する時も、X-Content-Type-Options: nosniff を付与する", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    const got = await app.request(`${origin}${episodeAudioPath("ep-1")}`, {}, emptyEnv);

    // Then: secureHeaders() が全route共通で効く
    expect(got.headers.get("X-Content-Type-Options")).toBe("nosniff");
  });

  it("一覧 GET が成功する時、requestId header を付与する", async () => {
    // Given: Composition が契約どおりの一覧を返す
    vi.mocked(listEpisodesController).mockResolvedValue(validList);

    // When: 一覧 path へ GET する
    const got = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: requestLoggingMiddleware が発行した requestId が response header に乗る
    expect(got.headers.get(requestIdHeaderName)).toBeTruthy();
  });

  it("音声 GET の Controller が NotFoundError を throw する時も、onError 側の requestId が開始ログと一致する", async () => {
    // Given: Domain 不在を写した External Error
    vi.mocked(getAudioController).mockRejectedValue(new NotFoundError("エピソードが無い"));

    // When: 音声 path へ GET する
    await app.request(`${origin}${episodeAudioPath("missing")}`, {}, emptyEnv);

    // Then: requestLoggingMiddleware の開始ログと onError 経由の error ログが同じ requestId を共有する
    const startCall = logSpy.mock.calls.find(([payload]) => payload.event === "request_start");
    const errorCall = errorSpy.mock.calls[0]?.[0];
    expect(startCall?.[0].requestId).toBe(errorCall.requestId);
  });

  it("runtime config の内部 Error を configuration_error へ変換し、診断を cause へ残す", async () => {
    // Given: Composition Root が設定不足を内部 Error として throw する
    vi.mocked(createPlaybackControllers).mockImplementationOnce(() => {
      throw new PlaybackRuntimeConfigError("EPISODES（R2 binding）が未設定です");
    });

    // When: 一覧 path へ GET する
    const got = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: HTTP boundary が 500 と契約 code へ変換する
    expect(got.status).toBe(500);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "configuration_error" });
    expect(listEpisodesController).not.toHaveBeenCalled();
    expect(errorSpy).toHaveBeenCalledTimes(1);
    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "ConfigurationError",
        message: "設定を確認できません",
        cause: {
          name: "PlaybackRuntimeConfigError",
          message: "EPISODES（R2 binding）が未設定です",
        },
      }),
    );
  });
});
