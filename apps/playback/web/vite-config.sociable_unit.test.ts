import { afterAll, afterEach, describe, expect, it, vi } from "vitest";
import {
  episodeAudioContentType,
  episodeAudioPath,
  ErrorResponseSchema,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../contracts/index.ts";
import { createFakeEpisodeAudioBytes } from "../worker/src/test/fixtures/audio-bytes.ts";
import { validEpisodeItem } from "../worker/src/controllers/fake-use-cases.ts";
import { createDummyBackendMiddleware } from "./vite.config.ts";

const origin = "http://localhost";

const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

afterEach(() => {
  errorSpy.mockClear();
});

afterAll(() => {
  errorSpy.mockRestore();
});

describe("createDummyBackendMiddleware", () => {
  it("一覧 path へ GET する時、dummy backend 由来の ListEpisodesResponse を 200 で返す", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 一覧 path へ GET する
    const got = await handle(new Request(`${origin}${listEpisodesPath}`));

    // Then: fake-use-cases 由来の schema 準拠 body が 200 で返る
    expect(got.status).toBe(200);
    const body: unknown = await got.json();
    expect(ListEpisodesResponseSchema.safeParse(body).success).toBe(true);
  });

  it("音声 path へ GET する時、契約の Content-Type で dummy backend 由来の byte を返す", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 音声 path へ GET する
    const got = await handle(new Request(`${origin}${episodeAudioPath("ep-1")}`));

    // Then: fake-use-cases の byte がそのまま配線される
    expect(got.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await got.arrayBuffer());
    const expected = createFakeEpisodeAudioBytes(validEpisodeItem.durationSec);
    expect(bytes.byteLength).toBe(expected.byteLength);
    expect(bytes[0]).toBe(0x52);
  });

  it("契約に無い path へ GET する時、400 と validation_error を返す", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 契約に無い path へ GET する
    const got = await handle(new Request(`${origin}/unknown`));

    // Then: worker/src/routes/app.ts の notFound 契約と同じ 400 validation_error を返す
    expect(got.status).toBe(400);
    const body: unknown = await got.json();
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(body).toEqual({ code: "validation_error" });
  });
});
