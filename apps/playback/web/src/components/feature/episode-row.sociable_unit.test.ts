import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeRow } from "./episode-row.tsx";

const episode: EpisodeData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [
      { title: "小題A", preface: "前A", detail: "詳A", startSec: 0 },
      { title: "小題B", preface: "前B", detail: "詳B", startSec: 30 },
    ],
    closing: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

describe("EpisodeRow", () => {
  it("root に episode-row class を付けて描画する", () => {
    // Given: episode と handler
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeRow, {
        episode,
        episodeCount: 1,
        episodeIndex: 0,
        isSelected: false,
        isPlaying: false,
        onSelect: vi.fn(),
        onPlay: vi.fn(),
        onStop: vi.fn(),
      }),
    );

    // Then: episode-row class が root にある
    expect(container.firstElementChild?.className).toBe("episode-row");
    expect(container.querySelector("[data-episode-title]")?.textContent).toBe("1.　題1");
  });

  it("再生中は停止ボタンを描画し、クリックで onStop を呼ぶ", () => {
    // Given: isPlaying=true と handler
    const onStop = vi.fn();
    const { container } = render(
      createElement(EpisodeRow, {
        episode,
        episodeCount: 1,
        episodeIndex: 0,
        isSelected: true,
        isPlaying: true,
        onSelect: vi.fn(),
        onPlay: vi.fn(),
        onStop,
      }),
    );

    // When: 再生/停止 button をクリックする
    const article = container.querySelector("article");
    expect(article?.getAttribute("data-selected")).toBe("true");
    expect(article?.getAttribute("data-playing")).toBe("true");
    const buttons = container.querySelectorAll("button");
    expect(buttons[1]?.textContent).toBe("停止");
    buttons[1]?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onStop が呼ばれる
    expect(onStop).toHaveBeenCalled();
  });

  it("非再生中は再生ボタンを描画し、クリックで onPlay を呼ぶ", () => {
    // Given: isPlaying=false と onPlay spy
    const onPlay = vi.fn();
    const { container } = render(
      createElement(EpisodeRow, {
        episode,
        episodeCount: 1,
        episodeIndex: 0,
        isSelected: false,
        isPlaying: false,
        onSelect: vi.fn(),
        onPlay,
        onStop: vi.fn(),
      }),
    );

    // When: 再生 button をクリックする
    const buttons = container.querySelectorAll("button");
    expect(buttons[1]?.textContent).toBe("再生");
    buttons[1]?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onPlay が episodeId 付きで呼ばれる
    expect(onPlay).toHaveBeenCalledWith("ep-1");
  });

  it("select button クリックで onSelect を呼ぶ", () => {
    // Given: onSelect spy
    const onSelect = vi.fn();
    const { container } = render(
      createElement(EpisodeRow, {
        episode,
        episodeCount: 1,
        episodeIndex: 0,
        isSelected: false,
        isPlaying: false,
        onSelect,
        onPlay: vi.fn(),
        onStop: vi.fn(),
      }),
    );

    // When: select button をクリックする
    container.querySelectorAll("button")[0]?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が episodeId 付きで呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });
});
