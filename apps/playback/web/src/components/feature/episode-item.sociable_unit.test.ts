import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeItem } from "./episode-item.tsx";

const episode: EpisodeData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: { text: "開始文", startSec: 0 },
    topics: [
      { title: "小題A", preface: "前A", detail: "詳A", startSec: 0 },
      { title: "小題B", preface: "前B", detail: "詳B", startSec: 30 },
    ],
    ending: { text: "終了文", startSec: 55 },
  },
  audioRef: "/episodes/ep-1/audio",
};

function renderItem(overrides: Partial<Parameters<typeof EpisodeItem>[0]> = {}) {
  return render(
    createElement(EpisodeItem, {
      episode,
      episodeCount: 1,
      episodeIndex: 0,
      isSelected: false,
      isActivePlayback: false,
      isPlaying: false,
      onSelect: vi.fn(),
      onPlay: vi.fn(),
      onStop: vi.fn(),
      onSeek: vi.fn(),
      ...overrides,
    }),
  );
}

describe("EpisodeItem", () => {
  it("root は article.episode-item で、常に EpisodeRow を含む", () => {
    // Given: 未選択の row
    // When: JSX として render する
    const { container } = renderItem();

    // Then: 枠は article.episode-item、中に行本体が居る
    const article = container.firstElementChild as HTMLElement;
    expect(article.tagName).toBe("ARTICLE");
    expect(article.className).toBe("episode-item");
    expect(article.querySelector(".episode-row")).not.toBeNull();
  });

  it("未選択なら EpisodeManuscript を描画しない・data-selected は false", () => {
    // Given: isSelected=false
    const { container } = renderItem({ isSelected: false });

    // Then: 原稿は出ず、data-selected は false
    expect(container.querySelector(".episode-manuscript")).toBeNull();
    expect(container.querySelector(".episode-item")?.getAttribute("data-selected")).toBe("false");
  });

  it("選択中は EpisodeManuscript を EpisodeRow の直後に描画し、data-selected を true にする", () => {
    // Given: isSelected=true
    const { container } = renderItem({ isSelected: true });

    // Then: data-selected="true"（横線は CSS 側の責務）。行本体の直後に原稿が来る
    const article = container.querySelector(".episode-item") as HTMLElement;
    expect(article.getAttribute("data-selected")).toBe("true");
    const children = Array.from(article.children);
    expect(children[0]?.classList.contains("episode-row")).toBe(true);
    expect(children[1]?.classList.contains("episode-manuscript")).toBe(true);
    expect(container.querySelector("[data-manuscript-opening]")?.textContent).toBe("開始文");
  });

  it("選択中に原稿の topic seek を押すと onSeek が startSec 付きで呼ばれる", () => {
    // Given: onSeek spy、選択中
    const onSeek = vi.fn();
    const { container } = renderItem({ isSelected: true, onSeek });

    // When: 2 番目の topic の seek button（startSec 30）を押す
    const seekButtons = container.querySelectorAll(".episode-manuscript [data-topic-start-sec]");
    seekButtons[1]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: onSeek(30)
    expect(onSeek).toHaveBeenCalledWith(30);
  });

  it("再生ボタンを押すと onPlay を呼び、select は起こさない", () => {
    // Given: onPlay / onSelect spy
    const onPlay = vi.fn();
    const onSelect = vi.fn();
    const { container } = renderItem({ onPlay, onSelect });

    // When: 行本体の再生 button を押す
    container
      .querySelector(".episode-row__play")
      ?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: onPlay だけ
    expect(onPlay).toHaveBeenCalledWith("ep-1");
    expect(onSelect).not.toHaveBeenCalled();
  });
});
