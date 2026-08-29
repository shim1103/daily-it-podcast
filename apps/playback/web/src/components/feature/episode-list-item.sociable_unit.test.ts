import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeListItemData } from "../../view-models/episode-list-view-model.ts";
import { EpisodeListItem } from "./episode-list-item.tsx";

const episode: EpisodeListItemData = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  topics: [{ title: "小題" }],
};

describe("EpisodeListItem", () => {
  it("date をスラッシュ形式・durationSec を mm:ss で描画し、episodeId は出さない", () => {
    // Given: EpisodeListItem 1件（wire 形式の date / durationSec）
    // When: JSX として render する
    const { container } = render(createElement(EpisodeListItem, { episode, onSelect: vi.fn() }));

    // Then: 表示用に整形された日付・尺と title が出て、episodeId は可視 text に無い
    expect(container.querySelector("[data-episode-id]")).toBeNull();
    expect(container.textContent).not.toContain("ep-1");
    expect(container.querySelector("[data-episode-date]")?.textContent).toBe("2026/08/17");
    expect(container.querySelector("[data-episode-title]")?.textContent).toBe("題1");
    expect(container.querySelector("[data-episode-duration-sec]")?.textContent).toBe("01:00");
  });

  it("クリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: onSelect の spy を渡した component
    const onSelect = vi.fn();
    const { container } = render(createElement(EpisodeListItem, { episode, onSelect }));

    // When: button をクリックする
    container.querySelector("button")?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が episodeId 付きで呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });
});
