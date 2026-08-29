import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { EpisodeRow, type EpisodeRowProps } from "./episode-row.tsx";

const listEpisode: EpisodeRowProps["episode"] = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  topics: [{ title: "小題A" }, { title: "小題B" }],
  audioRef: "/episodes/ep-1/audio",
};

function renderEpisodeRow(overrides: Partial<EpisodeRowProps> = {}) {
  return render(
    createElement(EpisodeRow, {
      episode: listEpisode,
      episodeCount: 1,
      episodeIndex: 0,
      onSelect: vi.fn(),
      onPlay: vi.fn(),
      ...overrides,
    }),
  );
}

describe("EpisodeRow", () => {
  it("date をスラッシュ形式・durationSec を mm:ss で描画し、episodeId は出さない", () => {
    // Given: EpisodeRow 1件
    const { container } = renderEpisodeRow();

    // Then: 表示用に整形された日付・尺と title が出て、episodeId は可視 text に無い
    expect(container.querySelector("[data-episode-id]")).toBeNull();
    expect(container.textContent).not.toContain("ep-1");
    expect(container.querySelector("[data-episode-date]")?.textContent).toBe("2026/08/17");
    expect(container.querySelector("[data-episode-title]")?.textContent).toBe("1.　題1");
    expect(container.querySelector(".episode-play-button__duration")?.textContent).toBe("01:00");
  });

  it("topics[].title を / 区切りで title の下に描画する", () => {
    // Given: topics を2件持つ episode
    const { container } = renderEpisodeRow({ episodeCount: 3, episodeIndex: 1 });

    // Then: topics 行が title の下に出る
    expect(container.querySelector("[data-episode-topics]")?.textContent).toBe("小題A / 小題B");
  });

  it("topics が空の時は topics 行を描画しない", () => {
    // Given: topics が空の episode
    const { container } = renderEpisodeRow({
      episode: { ...listEpisode, topics: [] },
    });

    // Then: topics 行が無い
    expect(container.querySelector("[data-episode-topics]")).toBeNull();
  });

  it("行 button をクリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: onSelect の spy
    const onSelect = vi.fn();
    const { container } = renderEpisodeRow({ onSelect });

    // When: 行 button をクリックする
    container
      .querySelector(".episode-row__hit")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が episodeId 付きで呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });

  it("play pill を click すると onPlay を呼ぶ", () => {
    // Given: onPlay の spy
    const onPlay = vi.fn();
    const { container } = renderEpisodeRow({ onPlay });

    // When: play pill を click する
    container
      .querySelector(".episode-play-button")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onPlay が呼ばれる
    expect(onPlay).toHaveBeenCalledTimes(1);
  });
});
