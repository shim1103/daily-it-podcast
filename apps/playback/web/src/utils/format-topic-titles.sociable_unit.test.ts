import { describe, expect, it } from "vitest";
import { formatTopicTitles } from "./format-topic-titles.ts";

describe("formatTopicTitles", () => {
  it("topics[].title を出現順のまま / 区切りで連結する", () => {
    // Given: title を2件持つ topics
    const topics = [{ title: "型狭め" }, { title: "zod 境界" }];

    // When: 整形する
    const got = formatTopicTitles(topics);

    // Then: / 区切りの1行になる
    expect(got).toBe("型狭め / zod 境界");
  });

  it("topics が空の時は空文字を返す", () => {
    // Given: 空配列
    // When: 整形する
    const got = formatTopicTitles([]);

    // Then: 空文字
    expect(got).toBe("");
  });
});
