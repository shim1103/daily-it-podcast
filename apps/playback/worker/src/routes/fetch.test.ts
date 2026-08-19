import { afterAll, afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorResponseSchema,
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  NotFoundError,
  UnavailableError,
  ValidationError,
  episodeAudioContentType,
  episodeAudioPath,
  episodePath,
  listEpisodesPath,
} from "../../../contracts/index.ts";

const listEpisodesController = vi.fn();
const getEpisodeController = vi.fn();
const getEpisodeAudioController = vi.fn();

vi.mock("../composition/root.ts", () => ({
  createPlaybackControllers: vi.fn(() => ({
    kind: "ready",
    controllers: {
      listEpisodesController,
      getEpisodeController,
      getEpisodeAudioController,
    },
  })),
}));

import { createPlaybackControllers } from "../composition/root.ts";
import { fetch as handleFetch } from "./fetch.ts";

const origin = "http://example.test";
const emptyEnv = {};

const validList = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題",
      durationSec: 60,
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

describe("fetch", () => {
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
    await handleFetch(new Request(`${origin}${listEpisodesPath}`), driveEnv);

    // Then: 渡された env がそのまま Composition Root に渡る
    expect(createPlaybackControllers).toHaveBeenCalledWith(driveEnv);
  });

  it("一覧 GET が成功する時、ListEpisodesResponse schema を満たす JSON を 200 で返す", async () => {
    // Given: Composition が契約どおりの一覧を返す
    vi.mocked(listEpisodesController).mockResolvedValue(validList);

    // When: 一覧 path へ GET する
    const got = await handleFetch(new Request(`${origin}${listEpisodesPath}`), emptyEnv);

    // Then: 200 と契約 schema
    expect(got.status).toBe(200);
    const body: unknown = await got.json();
    expect(ListEpisodesResponseSchema.safeParse(body).success).toBe(true);
  });

  it("1件 JSON GET が成功する時、GetEpisodeResponse schema を満たし audioRef がある", async () => {
    // Given: Composition が契約どおりの 1件を返す
    vi.mocked(getEpisodeController).mockResolvedValue(validGet);

    // When: 1件 path へ GET する
    const got = await handleFetch(new Request(`${origin}${episodePath("ep-1")}`), emptyEnv);

    // Then: 200 と契約 schema
    expect(got.status).toBe(200);
    const body: unknown = await got.json();
    expect(GetEpisodeResponseSchema.safeParse(body).success).toBe(true);
  });

  it("1件 JSON の path 段を unknown の episodeId として Controller に渡す", async () => {
    // Given: Composition が 1件を返す
    vi.mocked(getEpisodeController).mockResolvedValue(validGet);

    // When: 1件 path へ GET する
    await handleFetch(new Request(`${origin}${episodePath("ep-1")}`), emptyEnv);

    // Then: schema parse せず unknown で渡す
    expect(getEpisodeController).toHaveBeenCalledWith({ episodeId: "ep-1" });
  });

  it("空の path 段を unknown の episodeId として Controller に渡す", async () => {
    // Given: 空 episodeId は Controller が ValidationError にする
    vi.mocked(getEpisodeController).mockRejectedValue(new ValidationError("入力が契約に不適合"));

    // When: 末尾スラッシュだけの 1件 path へ GET する
    await handleFetch(new Request(`${origin}${listEpisodesPath}/`), emptyEnv);

    // Then: schema parse せず空文字を unknown で渡す
    expect(getEpisodeController).toHaveBeenCalledWith({ episodeId: "" });
  });

  it("Controller が ValidationError を throw する時、400 と validation_error を返す", async () => {
    // Given: 空 episodeId 相当の External Error
    vi.mocked(getEpisodeController).mockRejectedValue(new ValidationError("入力が契約に不適合"));

    // When: 1件 path へ GET する
    const got = await handleFetch(new Request(`${origin}${episodePath("ep-1")}`), emptyEnv);

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
    const got = await handleFetch(new Request(`${origin}${episodePath("missing")}`), emptyEnv);

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
    const got = await handleFetch(new Request(`${origin}${episodePath("ep-1")}`), emptyEnv);

    // Then: 503 と契約 code のみ
    expect(got.status).toBe(503);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "unavailable" });
  });

  it("音声 GET が成功する時、契約の Content-Type で byte を返す", async () => {
    // Given: Composition が音声 byte を返す
    vi.mocked(getEpisodeAudioController).mockResolvedValue(validAudioBytes);

    // When: 音声 path へ GET する
    const got = await handleFetch(new Request(`${origin}${episodeAudioPath("ep-1")}`), emptyEnv);

    // Then: JSON ではなく契約 Content-Type の byte
    expect(got.status).toBe(200);
    expect(got.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await got.arrayBuffer());
    expect(bytes).toEqual(validAudioBytes);
  });

  it("method または path が契約に無い時、400 と validation_error を返す", async () => {
    // Given: 契約に無い path
    const request = new Request(`${origin}/unknown`);

    // When: GET する
    const got = await handleFetch(request, emptyEnv);

    // Then: 未一致を episode_not_found に畳まない
    expect(got.status).toBe(400);
    const body: unknown = await got.json();
    expect(body).toEqual({ code: "validation_error" });
  });

  it("POST の一覧 path は 400 と validation_error を返す", async () => {
    // Given: 契約 method ではない POST
    const request = new Request(`${origin}${listEpisodesPath}`, {
      method: "POST",
    });

    // When: 送る
    const got = await handleFetch(request, emptyEnv);

    // Then: 未一致は validation_error
    expect(got.status).toBe(400);
    const body: unknown = await got.json();
    expect(body).toEqual({ code: "validation_error" });
  });

  it("Composition Root が misconfigured を返す時、503 と unavailable を返す", async () => {
    // Given: Drive env が一部だけ欠けた状態（無言で Fake へは逃げない）
    vi.mocked(createPlaybackControllers).mockReturnValueOnce({
      kind: "misconfigured",
      missing: ["DRIVE_FOLDER_ID"],
    });

    // When: 一覧 path へ GET する
    const got = await handleFetch(new Request(`${origin}${listEpisodesPath}`), emptyEnv);

    // Then: 503 と契約 code のみ。Controller は呼ばれない
    expect(got.status).toBe(503);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "unavailable" });
    expect(listEpisodesController).not.toHaveBeenCalled();
  });

  it("ValidationError を structured payload で log し Error object 自体は渡さない", async () => {
    // Given: ValidationError
    const thrown = new ValidationError("入力が契約に不適合", {
      cause: new Error("zod"),
    });
    vi.mocked(getEpisodeController).mockRejectedValue(thrown);

    // When: 1件 path へ GET する
    await handleFetch(new Request(`${origin}${episodePath("ep-1")}`), emptyEnv);

    // Then: name / message / stack / cause / requestId を持ち Error ではない
    expect(errorSpy).toHaveBeenCalledTimes(1);
    const payload = errorSpy.mock.calls[0]?.[0];
    expect(payload).not.toBeInstanceOf(Error);
    expect(payload).toEqual(
      expect.objectContaining({
        name: "ValidationError",
        message: thrown.message,
        stack: thrown.stack,
        requestId: expect.any(String),
        cause: {
          name: "Error",
          message: "zod",
        },
      }),
    );
  });
});
