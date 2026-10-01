// @vitest-environment node
/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers R2 binding
 * Double: 本番 remote R2 は使わない（remoteBindings: false）
 *
 * 手書きの binding 型 `R2BucketBinding` と実 binding の構造的一致（`list`・`get` の戻り形、`arrayBuffer()` が
 * adapter の形状検査を通ること）を、実 binding を刺して守る。`R2Bucket` が型 scope に無く tsc が検査しないため
 * （Decision 2026-09-15T14-28-12）。adapter の分岐（不正形状・欠落・失敗の写像）は sociable unit が所有する。
 *
 * @require createLocalR2Binding が実 proxy を起動する
 * @ensure R2EpisodeRepository が実 R2 binding に対して原稿一覧と mp3 取得の観測可能な振る舞いを満たす
 */
import { afterAll, beforeAll, beforeEach, describe, expect, it } from "vitest";
import { R2EpisodeRepository } from "../../worker/src/infrastructure/r2/r2-episode-repository.ts";
import {
  type LocalR2BindingHandle,
  createLocalR2Binding,
} from "../support/create-local-r2-binding.ts";

// why: `R2BucketBinding` は読み出し面（get・list）だけを持つ。実 binding の put・delete は test の fixture 投入と
//   後始末にだけ使うため、この test 内に閉じた型で cast する
type R2BucketWithWrites = {
  put(key: string, value: string | Uint8Array): Promise<unknown>;
  delete(key: string): Promise<void>;
  list(): Promise<{ objects: Array<{ key: string }> }>;
};

// why: 起動に失敗すると handle は未代入のまま afterAll に来る。dispose の TypeError で元の失敗原因を隠さない
let handle: LocalR2BindingHandle | undefined;
let bucket: R2BucketWithWrites;
let repository: R2EpisodeRepository;

beforeAll(async () => {
  handle = await createLocalR2Binding();
  bucket = handle.bucket as unknown as R2BucketWithWrites;
  repository = new R2EpisodeRepository({ bucket: handle.bucket });
});

afterAll(async () => {
  await handle?.dispose();
});

beforeEach(async () => {
  const { objects } = await bucket.list();
  for (const object of objects) {
    await bucket.delete(object.key);
  }
});

describe("R2EpisodeRepository on local R2", () => {
  describe("listManuscripts", () => {
    it("returns_stem_and_decoded_json_for_each_manuscript_object_when_bucket_holds_manuscripts", async () => {
      // Given: 配置契約の拡張子（.json）で put された原稿 2 件（日本語を含む）
      await bucket.put("ep-1.json", JSON.stringify({ title: "第一回" }));
      await bucket.put("ep-2.json", JSON.stringify({ title: "第二回" }));

      // When: 原稿一覧を取得する
      const got = await repository.listManuscripts();

      // Then: stem と decode 済み json が実 binding の get 戻り値から揃う
      expect([...got].sort((a, b) => a.stem.localeCompare(b.stem))).toEqual([
        { stem: "ep-1", json: { title: "第一回" } },
        { stem: "ep-2", json: { title: "第二回" } },
      ]);
    });
  });

  describe("getAudio", () => {
    it("returns_put_bytes_as_uint8array_when_mp3_object_exists", async () => {
      // Given: UTF-8 として不正な byte を含む mp3 object
      const bytes = new Uint8Array([0xff, 0x00, 0x49, 0x44, 0x33, 0xfe]);
      await bucket.put("ep-1.mp3", bytes);

      // When: episodeId で音声を取得する
      const got = await repository.getAudio("ep-1");

      // Then: put した byte 列が Uint8Array でそのまま読み戻せる
      expect(got).toBeInstanceOf(Uint8Array);
      expect(got).toEqual(bytes);
    });

    it("returns_undefined_when_mp3_object_does_not_exist", async () => {
      // Given: mp3 の無い bucket

      // When: 存在しない episodeId で音声を取得する
      const got = await repository.getAudio("ep-unknown");

      // Then: 実 binding の get が返す null が undefined になる
      expect(got).toBeUndefined();
    });
  });
});
