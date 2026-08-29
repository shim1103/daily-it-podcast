import { describe, expect, it } from "vitest";
import { formatNumberedTopicTitle } from "./format-numbered-topic-title.ts";

describe("formatNumberedTopicTitle", () => {
  it("topicIndex 0 始まりを 1 始まりの番号付き title にする", () => {
    // Given: 2 番目の topic（index 1）
    const got = formatNumberedTopicTitle(1, "小題");

    // Then: 2. 小題
    expect(got).toBe("2. 小題");
  });
});
