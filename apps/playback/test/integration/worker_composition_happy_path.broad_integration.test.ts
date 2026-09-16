// @vitest-environment node
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorResponseSchema,
  episodeAudioContentType,
  episodeAudioPath,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../../contracts/index.ts";
import type { R2BucketBinding } from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";
import { validAudioBytes } from "../../worker/src/test/fixtures/audio-bytes.ts";
import workerEntry from "../../worker/src/worker-entry.ts";

/**
 * scope: Broad Integration
 * real: Worker entry・route・Composition Root・Controller・UseCase・R2EpisodeRepository
 * double: R2 binding（`R2BucketBinding` の in-memory 実装）。真の Cloudflare R2 へは行かない
 * precondition: 本番route（options.mode: "r2" 固定）が R2 repository を選ぶ
 * postcondition: list / get audio の成功応答が入口から見える。代表の R2 失敗は 503 unavailable
 * invariant: PlaybackUseCaseOverrides で use case 直差ししない
 */

const episodeId = "bi-ep-1";

const manuscriptJson = {
  episodeId,
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: { text: "開始", startSec: 0 },
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    ending: { text: "終了", startSec: 55 },
  },
};

function bodyOf(bytes: Uint8Array): { arrayBuffer(): Promise<ArrayBuffer> } {
  return {
    async arrayBuffer() {
      const buffer = new ArrayBuffer(bytes.byteLength);
      new Uint8Array(buffer).set(bytes);
      return buffer;
    },
  };
}

function createBucket(overrides: Partial<R2BucketBinding> = {}): R2BucketBinding {
  return {
    get: async () => null,
    list: async () => ({ objects: [] }),
    ...overrides,
  };
}

function createHappyBucket(): R2BucketBinding {
  return createBucket({
    list: async () => ({ objects: [{ key: `${episodeId}.json` }, { key: `${episodeId}.mp3` }] }),
    get: async (key) => {
      if (key === `${episodeId}.json`) {
        return bodyOf(new TextEncoder().encode(JSON.stringify(manuscriptJson)));
      }
      if (key === `${episodeId}.mp3`) {
        return bodyOf(validAudioBytes);
      }
      return null;
    },
  });
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe("Playback Worker composition happy path", () => {
  it("returns 200 list episodes through Worker entry when R2 binding has manuscripts", async () => {
    // Given: R2 binding double が対象 episode の json を持つ
    const env = { EPISODES: createHappyBucket() };
    const request = new Request(`https://worker.example${listEpisodesPath}`);

    // When: Worker HTTP 入口へ一覧 GET
    const response = await workerEntry.fetch(request, env);

    // Then: 入口から list 成功が見える（下位 mapping の全 field 一致はしない）
    expect(response.status).toBe(200);
    const body: unknown = await response.json();
    const parsed = ListEpisodesResponseSchema.safeParse(body);
    expect(parsed.success).toBe(true);
    if (!parsed.success) {
      return;
    }
    expect(parsed.data.episodes).toHaveLength(1);
    expect(parsed.data.episodes[0]?.body.opening).toEqual({ text: "開始", startSec: 0 });
  });

  it("returns 200 audio bytes through Worker entry when R2 binding has the mp3", async () => {
    // Given: R2 binding double が対象 episode の mp3 を持つ
    const env = { EPISODES: createHappyBucket() };
    const request = new Request(`https://worker.example${episodeAudioPath(episodeId)}`);

    // When: Worker HTTP 入口へ音声 GET
    const response = await workerEntry.fetch(request, env);

    // Then: 入口から audio 成功が見える（bytes 完全一致はしない）
    expect(response.status).toBe(200);
    expect(response.headers.get("Content-Type")).toBe(episodeAudioContentType);
    const bytes = new Uint8Array(await response.arrayBuffer());
    expect(bytes.byteLength).toBeGreaterThan(0);
  });

  it("returns 503 unavailable when R2 list fails through composition", async () => {
    // Given: R2 binding は揃うが list I/O が失敗する（合成で初めて見える error 伝播の代表）
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    const env = {
      EPISODES: createBucket({
        list: async () => {
          throw new Error("network");
        },
      }),
    };
    const request = new Request(`https://worker.example${listEpisodesPath}`);

    // When: Worker HTTP 入口へ一覧 GET
    const response = await workerEntry.fetch(request, env);

    // Then: 入口から unavailable が見える（config 不足 BI と重複しない）
    expect(response.status).toBe(503);
    const body: unknown = await response.json();
    expect(body).toEqual({ code: "unavailable" });
    expect(ErrorResponseSchema.safeParse(body).success).toBe(true);
    expect(errorSpy).toHaveBeenCalledTimes(1);
  });
});
