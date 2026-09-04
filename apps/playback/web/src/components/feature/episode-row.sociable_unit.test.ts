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
    ending: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

function renderRow(overrides: Partial<Parameters<typeof EpisodeRow>[0]> = {}) {
  return render(
    createElement(EpisodeRow, {
      episode,
      episodeCount: 1,
      episodeIndex: 0,
      isSelected: false,
      isActivePlayback: false,
      isPlaying: false,
      onSelect: vi.fn(),
      onPlay: vi.fn(),
      onStop: vi.fn(),
      ...overrides,
    }),
  );
}

describe("EpisodeRow", () => {
  it("root に episode-row class を付け、select / play button を兄弟に置く", () => {
    // Given: episode と handler
    // When: JSX として render する
    const { container } = renderRow();

    // Then: root は div.episode-row。button の入れ子を作らず select / play を直下の兄弟に置く
    const root = container.firstElementChild as HTMLElement;
    expect(root.className).toBe("episode-row");
    const select = root.querySelector(":scope > .episode-row__select");
    const play = root.querySelector(":scope > .episode-row__play");
    expect(select).not.toBeNull();
    expect(play).not.toBeNull();
    expect(select?.querySelector("button")).toBeNull();
    expect(container.querySelector("[data-episode-title]")?.textContent).toBe("1.　題1");
  });

  it("select button クリックで onSelect を episodeId 付きで呼ぶ", () => {
    // Given: onSelect spy
    const onSelect = vi.fn();
    const { container } = renderRow({ onSelect });

    // When: select button をクリックする
    container
      .querySelector(".episode-row__select")
      ?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: onSelect が episodeId 付きで呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });

  it("再生対象でない row は再生 Icon（aria-label=再生）を右端に描画し、クリックで onPlay を呼ぶ（onSelect は呼ばない）", () => {
    // Given: onPlay / onSelect spy
    const onPlay = vi.fn();
    const onSelect = vi.fn();
    const { container } = renderRow({ isActivePlayback: false, onPlay, onSelect });

    // When: play button をクリックする
    const play = container.querySelector(".episode-row__play") as HTMLButtonElement;
    // 見た目は Icon、意味は aria-label が担う
    expect(play.getAttribute("aria-label")).toBe("再生");
    play.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: onPlay だけが呼ばれる（再生が優先。select は起こさない）
    expect(onPlay).toHaveBeenCalledWith("ep-1");
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("再生対象の row は phase を問わず停止 Icon（aria-label=停止）を描画し、クリックで onStop を呼ぶ", () => {
    // Given: isActivePlayback=true（loading 中でも）と onStop spy
    const onStop = vi.fn();
    const { container } = renderRow({ isActivePlayback: true, isPlaying: false, onStop });

    // Then: play button の aria-label は「停止」（isPlaying=false でも）
    const play = container.querySelector(".episode-row__play") as HTMLButtonElement;
    expect(play.getAttribute("aria-label")).toBe("停止");

    // When: play button をクリックする
    play.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: onStop が呼ばれる
    expect(onStop).toHaveBeenCalled();
  });

  it("isPlaying を data-playing として反映する（再生中の視覚強調用。停止トグルは isActivePlayback）", () => {
    // Given: isPlaying=true
    const { container } = renderRow({ isActivePlayback: true, isPlaying: true });

    // Then: data-playing="true"
    expect(container.querySelector(".episode-row")?.getAttribute("data-playing")).toBe("true");
  });

  it("isSelected を data-selected として反映する（横線・余白の描画は親の責務）", () => {
    // Given: isSelected=true
    const { container } = renderRow({ isSelected: true });

    // Then: data-selected="true" を持つ
    expect(container.querySelector(".episode-row")?.getAttribute("data-selected")).toBe("true");
  });

  it("topic が空なら topics 行を描画しない", () => {
    // Given: topics が空の episode
    const { container } = renderRow({
      episode: { ...episode, body: { ...episode.body, topics: [] } },
    });

    // Then: topics の LabeledText は出ない
    expect(container.querySelector("[data-episode-topics]")).toBeNull();
  });
});
