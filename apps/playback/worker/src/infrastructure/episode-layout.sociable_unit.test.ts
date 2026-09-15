import { describe, expect, it } from "vitest";
import { audioExtension, jsonExtension, stemOf } from "./episode-layout.ts";

/**
 * scope: Sociable Unit
 * real: stemOf / jsonExtension / audioExtension（`contracts/episode-layout.md` の配置契約実装）
 */

describe("episode-layout", () => {
  describe("jsonExtension / audioExtension", () => {
    it("配置契約どおりの拡張子を持つ", () => {
      // Given / When: 定数値
      // Then: `contracts/episode-layout.md` が定める配置と一致する
      expect(jsonExtension).toBe(".json");
      expect(audioExtension).toBe(".mp3");
    });
  });

  describe("stemOf", () => {
    it("指定拡張子で終わる名前から拡張子を除いた stem を返す", () => {
      // Given: jsonExtension で終わる名前
      // When: stem を取り出す
      const got = stemOf("ep-1.json", jsonExtension);

      // Then: 拡張子を除いた episodeId が返る
      expect(got).toBe("ep-1");
    });

    it("指定拡張子で終わらない名前の時、undefined を返す", () => {
      // Given: audioExtension で終わらない名前
      // When: stem を取り出す
      const got = stemOf("ep-1.json", audioExtension);

      // Then: 不一致は undefined
      expect(got).toBeUndefined();
    });
  });
});
