import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it } from "vitest";
import { AudioControls } from "./audio-controls.tsx";

describe("AudioControls", () => {
  it("audioRef を張った controls 付き audio を描画し、src は持たない（音源は hook が命令的に張る）", () => {
    // Given: audioRef
    const audioRef = createRef<HTMLAudioElement | null>();

    // When: JSX として render する
    const { container } = render(createElement(AudioControls, { audioRef, nowPlaying: null }));

    // Then: audio-controls class と controls 付き audio。src 属性は付かない
    expect(container.firstElementChild?.className).toBe("audio-controls");
    const audio = container.querySelector("audio");
    expect(audio).not.toBeNull();
    expect(audio?.hasAttribute("controls")).toBe(true);
    expect(audio?.hasAttribute("src")).toBe(false);
    // ref が要素へ張られている（hook が src / play を操作できる）
    expect(audioRef.current).toBe(audio);
  });

  it("nowPlaying=null なら見出し（日付・title）を描画しない", () => {
    // Given: 再生対象なし
    const audioRef = createRef<HTMLAudioElement | null>();

    // When: render する
    const { container } = render(createElement(AudioControls, { audioRef, nowPlaying: null }));

    // Then: 見出し要素が無い
    expect(container.querySelector("[data-now-playing]")).toBeNull();
  });

  it("nowPlaying があれば audio の上に日付と通し番号付き title を描画する", () => {
    // Given: 再生中 episode の見出し
    const audioRef = createRef<HTMLAudioElement | null>();
    const nowPlaying = { date: "2026/08/17", numberedTitle: "3.　題名" };

    // When: render する
    const { container } = render(createElement(AudioControls, { audioRef, nowPlaying }));

    // Then: 見出しは audio より前に置かれ、日付と title を持つ
    const heading = container.querySelector("[data-now-playing]");
    expect(heading).not.toBeNull();
    expect(heading?.querySelector("[data-now-playing-date]")?.textContent).toBe("2026/08/17");
    expect(heading?.querySelector("[data-now-playing-title]")?.textContent).toBe("3.　題名");
    const audio = container.querySelector("audio") as Element;
    const relation = heading?.compareDocumentPosition(audio) ?? 0;
    expect(relation & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});
