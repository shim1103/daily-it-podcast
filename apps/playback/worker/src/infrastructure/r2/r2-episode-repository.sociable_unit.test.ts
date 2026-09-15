import { describe, expect, it } from "vitest";
import { validAudioBytes } from "../../test/fixtures/audio-bytes.ts";
import { R2Error } from "./r2-error.ts";
import { R2EpisodeRepository, type R2BucketBinding } from "./r2-episode-repository.ts";

/**
 * scope: Sociable Unit
 * real: R2EpisodeRepository
 * double: R2BucketBinding（in-memory 実装）
 */

const manuscriptJson = { episodeId: "ep-1", title: "題" };

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

describe("R2EpisodeRepository", () => {
  describe("listManuscripts", () => {
    it("prefix 一覧の .json だけを stem 付き生 payload として返す", async () => {
      // Given: .json と .mp3 が混在する一覧
      const bucket = createBucket({
        list: async () => ({
          objects: [{ key: "ep-1.json" }, { key: "ep-1.mp3" }],
        }),
        get: async (key) => {
          if (key === "ep-1.json") {
            return bodyOf(new TextEncoder().encode(JSON.stringify(manuscriptJson)));
          }
          return null;
        },
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When: 一覧を取得する
      const got = await repository.listManuscripts();

      // Then: .json の stem だけが原稿として返る
      expect(got).toEqual([{ stem: "ep-1", json: manuscriptJson }]);
    });

    it("一覧がゼロ件の時、空配列を返す（null でない）", async () => {
      // Given: 空の bucket
      const repository = new R2EpisodeRepository({ bucket: createBucket() });

      // When / Then
      expect(await repository.listManuscripts()).toEqual([]);
    });

    it("json として parse できない本文は decode した生文字列をそのまま返す", async () => {
      // Given: 不正 JSON の本文
      const bucket = createBucket({
        list: async () => ({ objects: [{ key: "ep-1.json" }] }),
        get: async () => bodyOf(new TextEncoder().encode("not json")),
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When: 一覧を取得する
      const got = await repository.listManuscripts();

      // Then: schema 判定はせず生文字列のまま返す
      expect(got).toEqual([{ stem: "ep-1", json: "not json" }]);
    });

    it("list が例外を throw する時、R2Error を throw する", async () => {
      // Given: 一覧取得が失敗する bucket
      const bucket = createBucket({
        list: async () => {
          throw new Error("network");
        },
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then: storage I/O 失敗は R2Error
      await expect(repository.listManuscripts()).rejects.toBeInstanceOf(R2Error);
    });

    it("list 後の get が非 null だが不正形状の時、R2Error を throw する", async () => {
      // Given: get が arrayBuffer を持たない不正形状を返す bucket
      const bucket = createBucket({
        list: async () => ({ objects: [{ key: "ep-1.json" }] }),
        // biome-ignore lint/suspicious/noExplicitAny: 不正形状を意図的に注入する test 用 double
        get: async () => ({}) as any,
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then: 不正形状も R2Error
      await expect(repository.listManuscripts()).rejects.toBeInstanceOf(R2Error);
    });

    it("get が例外を throw する時、R2Error を throw する", async () => {
      // Given: json 本文の取得が失敗する bucket
      const bucket = createBucket({
        list: async () => ({ objects: [{ key: "ep-1.json" }] }),
        get: async () => {
          throw new Error("network");
        },
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then
      await expect(repository.listManuscripts()).rejects.toBeInstanceOf(R2Error);
    });

    it("list 直後に対象 object が消えて get が null を返す時、R2Error を throw する", async () => {
      // Given: list には載るが、get 時点では既に削除済みの object（list-then-lost）
      const bucket = createBucket({
        list: async () => ({ objects: [{ key: "ep-1.json" }] }),
        get: async () => null,
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then: 一貫性違反は undefined ではなく R2Error
      await expect(repository.listManuscripts()).rejects.toBeInstanceOf(R2Error);
    });

    it("json 本文の arrayBuffer 読み出しが失敗する時、R2Error を throw する", async () => {
      // Given: object body の bytes 読み出し自体が失敗する bucket
      const bucket = createBucket({
        list: async () => ({ objects: [{ key: "ep-1.json" }] }),
        get: async () => ({
          async arrayBuffer() {
            throw new Error("stream error");
          },
        }),
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then
      await expect(repository.listManuscripts()).rejects.toBeInstanceOf(R2Error);
    });
  });

  describe("getAudio", () => {
    it("対応 mp3 がある時、byte をそのまま返す", async () => {
      // Given: episodeId に対応する mp3 が get できる bucket
      const bucket = createBucket({
        get: async (key) => {
          if (key === "ep-1.mp3") {
            return bodyOf(validAudioBytes);
          }
          return null;
        },
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When: 音声を取得する
      const got = await repository.getAudio("ep-1");

      // Then: 格納 byte と一致する
      expect(got).toEqual(validAudioBytes);
    });

    it("対応 mp3 が無い時、undefined を返す（throw しない）", async () => {
      // Given: get が null を返す bucket
      const repository = new R2EpisodeRepository({ bucket: createBucket() });

      // When / Then
      expect(await repository.getAudio("missing")).toBeUndefined();
    });

    it("get が例外を throw する時、R2Error を throw する", async () => {
      // Given: 音声取得が失敗する bucket
      const bucket = createBucket({
        get: async () => {
          throw new Error("network");
        },
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then: storage I/O 失敗は R2Error
      await expect(repository.getAudio("ep-1")).rejects.toBeInstanceOf(R2Error);
    });

    it("get が非 null だが不正形状の時、R2Error を throw する", async () => {
      // Given: arrayBuffer を持たない不正形状を返す bucket
      const bucket = createBucket({
        // biome-ignore lint/suspicious/noExplicitAny: 不正形状を意図的に注入する test 用 double
        get: async () => ({}) as any,
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then
      await expect(repository.getAudio("ep-1")).rejects.toBeInstanceOf(R2Error);
    });

    it("mp3 の arrayBuffer 読み出しが失敗する時、R2Error を throw する", async () => {
      // Given: object body の bytes 読み出し自体が失敗する bucket
      const bucket = createBucket({
        get: async () => ({
          async arrayBuffer() {
            throw new Error("stream error");
          },
        }),
      });
      const repository = new R2EpisodeRepository({ bucket });

      // When / Then
      await expect(repository.getAudio("ep-1")).rejects.toBeInstanceOf(R2Error);
    });
  });

  it("Error message に bucket key の実値を含めない", async () => {
    // Given: 一覧取得が失敗する bucket
    const bucket = createBucket({
      list: async () => {
        throw new Error("network");
      },
    });
    const repository = new R2EpisodeRepository({ bucket });

    // When / Then: message は固定語のみで key 実値を含めない
    await expect(repository.listManuscripts()).rejects.toSatisfy((error: unknown) => {
      return error instanceof R2Error && !error.message.includes("ep-1");
    });
  });
});
