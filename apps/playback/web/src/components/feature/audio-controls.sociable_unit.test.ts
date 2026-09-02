import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it } from "vitest";
import { AudioControls } from "./audio-controls.tsx";

describe("AudioControls", () => {
  it("audioRef と audioSrc から controls 付き audio を描画する", () => {
    // Given: audioRef と audioSrc
    const audioRef = createRef<HTMLAudioElement | null>();

    // When: JSX として render する
    const { container } = render(
      createElement(AudioControls, {
        audioRef,
        audioSrc: "/episodes/ep-1/audio",
      }),
    );

    // Then: audio-controls class と controls・src 付き audio
    expect(container.firstElementChild?.className).toBe("audio-controls");
    const audio = container.querySelector("audio");
    expect(audio).not.toBeNull();
    expect(audio?.hasAttribute("controls")).toBe(true);
    expect(audio?.getAttribute("src")).toBe("/episodes/ep-1/audio");
  });
});
