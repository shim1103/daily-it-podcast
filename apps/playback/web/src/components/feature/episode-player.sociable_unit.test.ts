import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { EpisodePlayer } from "./episode-player.tsx";

describe("EpisodePlayer", () => {
  it("root に episode-player class を付ける", () => {
    // Given: baseUrl と audioRef
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodePlayer, {
        baseUrl: "https://example.com",
        audioRef: "/episodes/ep-1/audio",
      }),
    );

    // Then: player 容器の class が root にある（見た目は CSS 側の責務）
    expect(container.firstElementChild?.className).toBe("episode-player");
  });

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
