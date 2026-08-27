import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { EpisodePlayer } from "./episode-player.tsx";

describe("EpisodePlayer", () => {
  it("audioRef と baseUrl から <audio controls src> を組み立てる", () => {
    // Given: baseUrl と audioRef
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodePlayer, {
        baseUrl: "https://example.com",
        audioRef: "/episodes/ep-1/audio",
      }),
    );

    // Then: audio要素が controls・組み立て済み src を持つ
    const element = container.querySelector("audio");
    expect(element).not.toBeNull();
    expect(element?.hasAttribute("controls")).toBe(true);
    expect(element?.getAttribute("src")).toBe("https://example.com/episodes/ep-1/audio");
  });
});
