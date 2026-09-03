import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeListItemData } from "../../view-models/episode-list-view-model.legacy.ts";
import { EpisodeListItem } from "./episode-list-item.legacy.tsx";

const episode: EpisodeListItemData = {
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

describe("EpisodeListItem", () => {
  it("date をスラッシュ形式・durationSec を mm:ss で描画し、episodeId は出さない", () => {
    // Given: EpisodeListItem 1件（wire 形式の date / durationSec）
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeListItem, {
        episode,
        episodeCount: 1,
        episodeIndex: 0,
        onSelect: vi.fn(),
      }),
    );

    // Then: 表示用に整形された日付・尺と title が出て、episodeId は可視 text に無い
    expect(container.querySelector("[data-episode-id]")).toBeNull();
    expect(container.textContent).not.toContain("ep-1");
    expect(container.querySelector("[data-episode-date]")?.textContent).toBe("2026/08/17");
    expect(container.querySelector("[data-episode-title]")?.textContent).toBe("1.　題1");
    expect(container.querySelector("[data-episode-duration-sec]")?.textContent).toBe("01:00");
  });

  it("topics[].title を / 区切りで title の下に描画する", () => {
    // Given: topics を2件持つ episode
    const onSelect = vi.fn();
    const { container } = render(
      createElement(EpisodeListItem, {
        episode,
        episodeCount: 3,
        episodeIndex: 1,
        onSelect,
      }),
    );

    // When: （render 済み）
    // Then: topics 行が title の下に出る
    expect(container.querySelector("[data-episode-topics]")?.textContent).toBe("小題A / 小題B");
  });

  it("topics が空の時は topics 行を描画しない", () => {
    // Given: body.topics が空の episode（component 単体の表示分岐）
    const onSelect = vi.fn();
    const { container } = render(
      createElement(EpisodeListItem, {
        episode: { ...episode, body: { ...episode.body, topics: [] } },
        episodeCount: 1,
        episodeIndex: 0,
        onSelect,
      }),
    );

    // Then: topics 行が無い
    expect(container.querySelector("[data-episode-topics]")).toBeNull();
  });

  it("クリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: onSelect の spy を渡した component
    const onSelect = vi.fn();
    const { container } = render(
      createElement(EpisodeListItem, {
        episode,
        episodeCount: 3,
        episodeIndex: 1,
        onSelect,
      }),
    );

    // When: button をクリックする
    container.querySelector("button")?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が episodeId 付きで呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });
});
