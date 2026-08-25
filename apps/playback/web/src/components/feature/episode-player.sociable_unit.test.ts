import { describe, expect, it } from "vitest";
import { createEpisodePlayer } from "./episode-player.ts";

describe("createEpisodePlayer", () => {
  it("audioRef と baseUrl から <audio controls src> を組み立てる", () => {
    // Given: baseUrl と audioRef
    // When: component を作る
    const element = createEpisodePlayer("https://example.com", "/episodes/ep-1/audio");

    // Then: audio要素が controls・組み立て済み src を持つ
    expect(element.tagName).toBe("AUDIO");
    expect(element.hasAttribute("controls")).toBe(true);
    expect(element.getAttribute("src")).toBe("https://example.com/episodes/ep-1/audio");
  });
});
