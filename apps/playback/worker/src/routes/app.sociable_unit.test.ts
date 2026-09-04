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
import { app, createApp } from "./app.ts";
import { validAudioBytes } from "../test/fixtures/audio-bytes.ts";

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
        opening: "開始",
        topics: [
          {
            title: "題",
            preface: "前置き",
            detail: "詳細",
            startSec: 0,
          },
        ],
        ending: "終了",
      },
      audioRef: episodeAudioPath("ep-1"),
    },
  ],
};

const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

afterEach(() => {
  vi.mocked(listEpisodesController).mockReset();
  vi.mocked(getAudioController).mockReset();
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
        getAudio: vi.fn(),
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

  it("音声 GET の path param を unknown の episodeId として Controller に渡す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    await app.request(`${origin}${episodeAudioPath("ep-1")}`, {}, emptyEnv);

    // Then: schema parse せず unknown で渡す
    expect(getAudioController).toHaveBeenCalledWith({ episodeId: "ep-1" });
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
