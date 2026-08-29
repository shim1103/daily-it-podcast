import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it } from "vitest";
import { EpisodeAudio } from "./episode-audio.tsx";

describe("EpisodeAudio", () => {
  it("hidden audio 容器を返す", () => {
    // Given: src あり
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeAudio, { src: "https://example.com/episodes/ep-1/audio" }),
    );

    // Then: audio 容器の class が root にある
    expect(container.firstElementChild?.className).toBe("episode-audio");
  });

  it("src を <audio> に渡す", () => {
    // Given: src あり
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeAudio, { src: "https://example.com/episodes/ep-1/audio" }),
    );

    // Then: audio 要素が src を持つ
    const element = container.querySelector("audio");
    expect(element).not.toBeNull();
    expect(element?.getAttribute("src")).toBe("https://example.com/episodes/ep-1/audio");
  });

  it("src が undefined の時、audio に src 属性を付けない", () => {
    // Given: src なし
    // When: JSX として render する
    const { container } = render(createElement(EpisodeAudio, {}));

    // Then: src 属性が無い
    expect(container.querySelector("audio")?.getAttribute("src")).toBeNull();
  });

  it("ref を audio 要素へ転送する", () => {
    // Given: ref object
    const ref = createRef<HTMLAudioElement | null>();

    // When: JSX として render する
    render(createElement(EpisodeAudio, { ref, src: "https://example.com/a" }));

    // Then: ref が audio を指す
    expect(ref.current).not.toBeNull();
    expect(ref.current?.tagName).toBe("AUDIO");
  });
});
