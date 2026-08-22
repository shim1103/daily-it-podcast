import { describe, expect, it } from "vitest";
import {
  episodeAudioContentType,
  episodeAudioPath,
  episodePath,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../contracts/index.ts";
import {
  validAudioBytes,
  validGetEpisodeResponse,
} from "../worker/src/controllers/fake-use-cases.ts";
import { createDummyBackendMiddleware } from "./vite.config.ts";

const origin = "http://localhost";

describe("createDummyBackendMiddleware", () => {
  it("一覧 path へ GET する時、dummy backend 由来の ListEpisodesResponse を 200 で返す", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 一覧 path へ GET する
    const got = await handle(new Request(`${origin}${listEpisodesPath}`));

    // Then: fake-use-cases 由来の schema 準拠 body が 200 で返る
    expect(got?.status).toBe(200);
    const body: unknown = await got?.json();
    expect(ListEpisodesResponseSchema.safeParse(body).success).toBe(true);
  });

  it("1件 path へ GET する時、dummy backend 由来の title を含む body を返す", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 1件 path へ GET する
    const got = await handle(new Request(`${origin}${episodePath("ep-1")}`));

    // Then: fake-use-cases が持つ title がそのまま配線される
    const body = (await got?.json()) as { title?: string };
    expect(body.title).toBe(validGetEpisodeResponse.title);
  });

  it("音声 path へ GET する時、契約の Content-Type で dummy backend 由来の byte を返す", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 音声 path へ GET する
    const got = await handle(new Request(`${origin}${episodeAudioPath("ep-1")}`));

    // Then: fake-use-cases の byte がそのまま配線される
    expect(got?.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array((await got?.arrayBuffer()) ?? new ArrayBuffer(0));
    expect(bytes).toEqual(validAudioBytes);
  });

  it("契約に無い path へ GET する時、undefined を返し next handler へ委ねる", async () => {
    // Given: dummy backend middleware
    const handle = createDummyBackendMiddleware();

    // When: 契約に無い path へ GET する
    const got = await handle(new Request(`${origin}/unknown`));

    // Then: 自前で 404 等を作らず undefined を返す
    expect(got).toBeUndefined();
  });
});
