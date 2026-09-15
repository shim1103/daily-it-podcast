// @vitest-environment node
import { describe, expect, it } from "vitest";
import { R2Error } from "../../worker/src/infrastructure/r2/r2-error.ts";
import {
  R2EpisodeRepository,
  type R2BucketBinding,
  type R2ObjectBodyLike,
} from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";
import { validAudioBytes } from "../../worker/src/test/fixtures/audio-bytes.ts";

/**
 * scope: Narrow Integration
 * 実物境界: R2EpisodeRepository が呼ぶ binding contract（list/get）
 * Double: R2 は HTTP ではなく binding 契約のため、実 HTTP server ではなく `R2BucketBinding` の
 *   in-memory fake を使う。
 * @require fake は `contracts/episode-layout.md` の配置契約（`{episodeId}.json` / `{episodeId}.mp3`）
 *   に従い object を保持する
 * @ensure list / get 成功・音声欠落・I/O 失敗・形式不正の代表ケースが Port 契約 / R2Error へ写る
 * @invariant error message に bucket key の実値を含めない
 */

const manuscriptJson = {
  episodeId: "narrow-ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: { text: "開始", startSec: 0 },
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    ending: { text: "終了", startSec: 55 },
  },
};

function bodyOf(bytes: Uint8Array): R2ObjectBodyLike {
  return {
    async arrayBuffer() {
      const buffer = new ArrayBuffer(bytes.byteLength);
      new Uint8Array(buffer).set(bytes);
      return buffer;
    },
  };
}

/**
 * `R2BucketBinding` の in-memory fake。Map への格納・取り出しだけを行う真の外部境界の double。
 */
class InMemoryR2Bucket implements R2BucketBinding {
  private readonly objects = new Map<string, Uint8Array>();

  put(key: string, bytes: Uint8Array): void {
    this.objects.set(key, bytes);
  }

  async get(key: string): Promise<R2ObjectBodyLike | null> {
    const bytes = this.objects.get(key);
    if (bytes === undefined) {
      return null;
    }
    return bodyOf(bytes);
  }

  async list(): Promise<{ objects: Array<{ key: string }> }> {
    const objects = [...this.objects.keys()].map((key) => ({ key }));
    return { objects };
  }
}

function createRepository(bucket: R2BucketBinding): R2EpisodeRepository {
  return new R2EpisodeRepository({ bucket });
}

describe("R2EpisodeRepository Narrow Integration", () => {
  it("listManuscripts returns stem and raw json when list and get succeed via binding", async () => {
    // Given: 配置契約通りに json / mp3 を保持する bucket
    const bucket = new InMemoryR2Bucket();
    bucket.put("narrow-ep-1.json", new TextEncoder().encode(JSON.stringify(manuscriptJson)));
    bucket.put("narrow-ep-1.mp3", validAudioBytes);
    const repository = createRepository(bucket);

    // When: 一覧を取得する
    const got = await repository.listManuscripts();

    // Then: .json だけが生 payload として返る
    expect(got).toEqual([{ stem: "narrow-ep-1", json: manuscriptJson }]);
  });

  it("getAudio returns mp3 bytes when the paired object exists", async () => {
    // Given: episodeId に対応する mp3 を保持する bucket
    const bucket = new InMemoryR2Bucket();
    bucket.put("narrow-ep-1.mp3", validAudioBytes);
    const repository = createRepository(bucket);

    // When: 音声を取得する
    const got = await repository.getAudio("narrow-ep-1");

    // Then: 格納した byte と一致する
    expect(got).toEqual(validAudioBytes);
  });

  it("getAudio returns undefined without throwing when the mp3 is missing", async () => {
    // Given: json のみで mp3 を保持しない bucket（音声欠落）
    const bucket = new InMemoryR2Bucket();
    bucket.put("narrow-ep-1.json", new TextEncoder().encode(JSON.stringify(manuscriptJson)));
    const repository = createRepository(bucket);

    // When / Then: throw せず undefined
    expect(await repository.getAudio("narrow-ep-1")).toBeUndefined();
  });

  it("throws R2Error without key values when list fails", async () => {
    // Given: list が例外を throw する bucket
    const failingBucket: R2BucketBinding = {
      get: async () => null,
      list: async () => {
        throw new Error("network");
      },
    };
    const repository = createRepository(failingBucket);

    // When / Then: storage I/O 失敗は R2Error
    await expect(repository.listManuscripts()).rejects.toSatisfy((error: unknown) => {
      return error instanceof R2Error && !error.message.includes("narrow-ep-1");
    });
  });

  it("throws R2Error without key values when get fails", async () => {
    // Given: json 本文の get が例外を throw する bucket
    const failingBucket: R2BucketBinding = {
      get: async () => {
        throw new Error("network");
      },
      list: async () => ({ objects: [{ key: "narrow-ep-1.json" }] }),
    };
    const repository = createRepository(failingBucket);

    // When / Then
    await expect(repository.listManuscripts()).rejects.toSatisfy((error: unknown) => {
      return error instanceof R2Error && !error.message.includes("narrow-ep-1");
    });
  });

  it("throws R2Error when get resolves to a malformed shape", async () => {
    // Given: get が arrayBuffer を持たない不正形状を返す bucket
    const malformedBucket: R2BucketBinding = {
      // biome-ignore lint/suspicious/noExplicitAny: binding 契約違反を意図的に注入する test 用 double
      get: async () => ({}) as any,
      list: async () => ({ objects: [] }),
    };
    const repository = createRepository(malformedBucket);

    // When / Then: 非 null だが不正形状も R2Error
    await expect(repository.getAudio("narrow-ep-1")).rejects.toBeInstanceOf(R2Error);
  });

  it("returns the raw decoded string when the json object body is not valid JSON", async () => {
    // Given: 不正 JSON の本文を持つ bucket（schema 判定は use-case 側の責務）
    const bucket = new InMemoryR2Bucket();
    bucket.put("narrow-ep-1.json", new TextEncoder().encode("not json"));
    const repository = createRepository(bucket);

    // When: 一覧を取得する
    const got = await repository.listManuscripts();

    // Then: decode した生文字列のまま返る
    expect(got).toEqual([{ stem: "narrow-ep-1", json: "not json" }]);
  });
});
