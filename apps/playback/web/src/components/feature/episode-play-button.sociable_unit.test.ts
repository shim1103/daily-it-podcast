import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { EpisodePlayButton } from "./episode-play-button.tsx";

describe("EpisodePlayButton", () => {
  it("再生 pill と duration を描画する", () => {
    // Given: durationSec
    const { container } = render(
      createElement(EpisodePlayButton, { durationSec: 60, onPlay: vi.fn() }),
    );

    // Then: pill button と duration が描画される
    expect(container.querySelector(".episode-play-button")).not.toBeNull();
    expect(container.querySelector(".episode-play-button__duration")?.textContent).toBe("01:00");
    expect(container.querySelector("audio")).toBeNull();
  });

  it("click すると onPlay を呼ぶ", () => {
    // Given: onPlay の spy
    const onPlay = vi.fn();
    const { container } = render(
      createElement(EpisodePlayButton, { durationSec: 60, onPlay }),
    );

    // When: pill を click する
    container
      .querySelector(".episode-play-button")
      ?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: onPlay が呼ばれる
    expect(onPlay).toHaveBeenCalledTimes(1);
  });
});
