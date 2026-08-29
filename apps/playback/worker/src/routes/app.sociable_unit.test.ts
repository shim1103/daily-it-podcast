import { afterAll, afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorResponseSchema,
  episodeAudioContentType,
  episodeAudioPath,
  episodePath,
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  listEpisodesPath,
  NotFoundError,
  UnavailableError,
  ValidationError,
} from "../../../contracts/index.ts";

const listEpisodesController = vi.fn();
const getEpisodeController = vi.fn();
const getEpisodeAudioController = vi.fn();

vi.mock("../composition/root.ts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../composition/root.ts")>();
  return {
    ...actual,
    createPlaybackControllers: vi.fn(() => ({
      listEpisodesController,
      getEpisodeController,
      getEpisodeAudioController,
    })),
  };
});

import { createPlaybackControllers, PlaybackRuntimeConfigError } from "../composition/root.ts";
import { app, createApp } from "./app.ts";

const origin = "http://example.test";
const emptyEnv = {};

const validList = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題",
      durationSec: 60,
      topics: [{ title: "題" }],
    },
  ],
};

const validGet = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      {
        title: "題",
        preface: "前置き",
        detail: "詳細",
        startSec: 0,
      },
    ],
    closing: "終了",
  },
  audioRef: episodeAudioPath("ep-1"),
};

const validAudioBytes = new Uint8Array([
  0x52, 0x49, 0x46, 0x46, 0x04, 0x00, 0x00, 0x00, 0x57, 0x41, 0x56, 0x45,
]);

const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

afterEach(() => {
  vi.mocked(listEpisodesController).mockReset();
  vi.mocked(getEpisodeController).mockReset();
  vi.mocked(getEpisodeAudioController).mockReset();
  vi.mocked(createPlaybackControllers).mockClear();
  errorSpy.mockClear();
});

afterAll(() => {
  errorSpy.mockRestore();
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
        getEpisode: vi.fn(),
        getEpisodeAudio: vi.fn(),
      },
    };
    const devApp = createApp(overrides);

    // When: 一覧 path へ GET する
    await devApp.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: createPlaybackControllers へ override がそのまま渡る
    expect(createPlaybackControllers).toHaveBeenCalledWith(emptyEnv, undefined, overrides);
  });

  it("受け取った env をそのまま Composition Root へ渡して Controller を組み立てる", async () => {
    // Given: Drive の env を模した値
    vi.mocked(listEpisodesController).mockResolvedValue(validList);
    const driveEnv = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
      DRIVE_FOLDER_ID: "folder-id",
    };

    // When: 一覧 path へ GET する
    await app.request(`${origin}${listEpisodesPath}`, {}, driveEnv);

    // Then: 渡された env がそのまま Composition Root に渡る
    expect(createPlaybackControllers).toHaveBeenCalledWith(driveEnv, undefined, undefined);
  });

  it("一覧 GET が成功する時、ListEpisodesResponse schema を満たす JSON を 200 で返す", async () => {
    // Given: Composition が契約どおりの一覧を返す
    vi.mocked(listEpisodesController).mockResolvedValue(validList);

    // When: 一覧 path へ GET する
    const got = await app.request(`${origin}${listEpisodesPath}`, {}, emptyEnv);

    // Then: 200 と契約 schema
    expect(got.status).toBe(200);
    const body: unknown = await got.json();
    expect(ListEpisodesResponseSchema.safeParse(body).success).toBe(true);
  });

  it("1件 JSON GET が成功する時、GetEpisodeResponse schema を満たし audioRef がある", async () => {
    // Given: Composition が契約どおりの 1件を返す
    vi.mocked(getEpisodeController).mockResolvedValue(validGet);

    // When: 1件 path へ GET する
    const got = await app.request(`${origin}${episodePath("ep-1")}`, {}, emptyEnv);

    // Then: 200 と契約 schema
    expect(got.status).toBe(200);
    const body: unknown = await got.json();
    expect(GetEpisodeResponseSchema.safeParse(body).success).toBe(true);
  });

  it("1件 JSON の path param を unknown の episodeId として Controller に渡す", async () => {
    // Given: Composition が 1件を返す
    vi.mocked(getEpisodeController).mockResolvedValue(validGet);

    // When: 1件 path へ GET する
    await app.request(`${origin}${episodePath("ep-1")}`, {}, emptyEnv);

    // Then: schema parse せず unknown で渡す
    expect(getEpisodeController).toHaveBeenCalledWith({ episodeId: "ep-1" });
  });

  it("日本語 episodeId を decode して Controller に渡す", async () => {
    // Given: Composition が 1件を返す
    vi.mocked(getEpisodeController).mockResolvedValue(validGet);

    // When: 日本語 episodeId を含む 1件 path へ GET する
    await app.request(`${origin}${episodePath("エピソード1")}`, {}, emptyEnv);

    // Then: decode 済みの episodeId を unknown で渡す
    expect(getEpisodeController).toHaveBeenCalledWith({ episodeId: "エピソード1" });
  });

  it("decode 不可能な episodeId の時、decode を諦めてそのまま Controller に渡す", async () => {
    // Given: Composition が 1件を返す
    vi.mocked(getEpisodeController).mockResolvedValue(validGet);

    // When: 単独 % を含む decode 不可能な 1件 path へ GET する
    await app.request(`${origin}${listEpisodesPath}/100%25`, {}, emptyEnv);

    // Then: decodeURIComponent が一度で失敗する元の文字列をそのまま渡す
    expect(getEpisodeController).toHaveBeenCalledWith({ episodeId: "100%" });
  });

  it("Controller が ValidationError を throw する時、400 と validation_error を返す", async () => {
    // Given: 空 episodeId 相当の External Error
    vi.mocked(getEpisodeController).mockRejectedValue(new ValidationError("入力が契約に不適合"));

    // When: 1件 path へ GET する
    const got = await app.request(`${origin}${episodePath("ep-1")}`, {}, emptyEnv);

    // Then: 400 と契約 code のみ
    expect(got.status).toBe(400);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "validation_error" });
  });

  it("Controller が NotFoundError を throw する時、404 と episode_not_found を返す", async () => {
    // Given: Domain 不在を写した External Error
    vi.mocked(getEpisodeController).mockRejectedValue(new NotFoundError("エピソードが無い"));

    // When: 1件 path へ GET する
    const got = await app.request(`${origin}${episodePath("missing")}`, {}, emptyEnv);

    // Then: 404 と契約 code のみ
    expect(got.status).toBe(404);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "episode_not_found" });
  });

  it("Controller が UnavailableError を throw する時、503 と unavailable を返す", async () => {
    // Given: Infrastructure 失敗を写した External Error
    vi.mocked(getEpisodeController).mockRejectedValue(new UnavailableError("利用できない"));

    // When: 1件 path へ GET する
    const got = await app.request(`${origin}${episodePath("ep-1")}`, {}, emptyEnv);

    // Then: 503 と契約 code のみ
    expect(got.status).toBe(503);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "unavailable" });
  });

  it("未知の error は契約外として 500 と empty body を返す", async () => {
    // Given: External Error mapping に存在しない error
    const thrown = new Error("予期しない失敗");
    vi.mocked(getEpisodeController).mockRejectedValue(thrown);

    // When: 1件 path へ GET する
    const got = await app.request(`${origin}${episodePath("ep-1")}`, {}, emptyEnv);

    // Then: 契約 enum を捏造せず、500 と empty body を返して structured log する
    expect(got.status).toBe(500);
    expect(await got.text()).toBe("");
    expect(errorSpy).toHaveBeenCalledTimes(1);
    expect(errorSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "UnmappedError",
        message: thrown.message,
        requestId: expect.any(String),
      }),
    );
  });

  it("音声 GET が成功する時、契約の Content-Type で byte を返す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getEpisodeAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    const got = await app.request(`${origin}${episodeAudioPath("ep-1")}`, {}, emptyEnv);

    // Then: JSON ではなく契約 Content-Type の byte
    expect(got.status).toBe(200);
    expect(got.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await got.arrayBuffer());
    expect(bytes).toEqual(validAudioBytes);
  });

  it("音声 GET の path param を unknown の episodeId として Controller に渡す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getEpisodeAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    await app.request(`${origin}${episodeAudioPath("ep-1")}`, {}, emptyEnv);

    // Then: schema parse せず unknown で渡す
    expect(getEpisodeAudioController).toHaveBeenCalledWith({ episodeId: "ep-1" });
  });

  it("音声 GET の Controller が NotFoundError を throw する時、404 と episode_not_found を返す", async () => {
    // Given: Domain 不在を写した External Error
    vi.mocked(getEpisodeAudioController).mockRejectedValue(new NotFoundError("エピソードが無い"));

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

  it("runtime config の内部 Error を configuration_error へ変換し、診断を cause へ残す", async () => {
    // Given: Composition Root が設定不足を内部 Error として throw する
    vi.mocked(createPlaybackControllers).mockImplementationOnce(() => {
      throw new PlaybackRuntimeConfigError(
        "GOOGLE_OAUTH_CLIENT_SECRET が未設定です; DRIVE_FOLDER_ID が未設定です",
      );
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
          message: "GOOGLE_OAUTH_CLIENT_SECRET が未設定です; DRIVE_FOLDER_ID が未設定です",
        },
      }),
    );
  });
});
