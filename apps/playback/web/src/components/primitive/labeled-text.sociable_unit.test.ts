import { describe, expect, it } from "vitest";
import { createLabeledText } from "./labeled-text.ts";

describe("createLabeledText", () => {
  it("指定した tag に dataset 属性と textContent をそのまま設定する", () => {
    // Given: tag・dataset属性名・text
    // When: 要素を作る
    const element = createLabeledText({ tag: "h1", datasetKey: "episodeTitle", text: "題1" });

    // Then: 指定した tag・dataset・textContent を持つ
    expect(element.tagName).toBe("H1");
    expect(element.dataset.episodeTitle).toBe("");
    expect(element.textContent).toBe("題1");
  });
});
