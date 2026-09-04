import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it } from "vitest";
import { AudioControls } from "./audio-controls.tsx";

describe("AudioControls", () => {
  it("audioRef を張った controls 付き audio を描画し、src は持たない（音源は hook が命令的に張る）", () => {
    // Given: audioRef
    const audioRef = createRef<HTMLAudioElement | null>();

    // When: JSX として render する
    const { container } = render(createElement(AudioControls, { audioRef }));

    // Then: audio-controls class と controls 付き audio。src 属性は付かない
    expect(container.firstElementChild?.className).toBe("audio-controls");
    const audio = container.querySelector("audio");
    expect(audio).not.toBeNull();
    expect(audio?.hasAttribute("controls")).toBe(true);
    expect(audio?.hasAttribute("src")).toBe(false);
    // ref が要素へ張られている（hook が src / play を操作できる）
    expect(audioRef.current).toBe(audio);
  });
});
