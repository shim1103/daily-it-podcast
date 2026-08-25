import { describe, expect, it, vi } from "vitest";
import type { ListEpisodesResponse } from "../../../../contracts/index.ts";
import { createEpisodeListItem } from "./episode-list-item.ts";

type EpisodeListItem = ListEpisodesResponse["episodes"][number];

const episode: EpisodeListItem = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
};

describe("createEpisodeListItem", () => {
  it("episodeId・date・title・durationSec をそのまま描画する", () => {
    // Given: EpisodeListItem 1件
    // When: component を作る
    const element = createEpisodeListItem(episode, vi.fn());

    // Then: 各 field がそのまま描画される
    expect(element.querySelector("[data-episode-id]")?.textContent).toBe("ep-1");
    expect(element.querySelector("[data-episode-date]")?.textContent).toBe("2026-08-17");
    expect(element.querySelector("[data-episode-title]")?.textContent).toBe("題1");
    expect(element.querySelector("[data-episode-duration-sec]")?.textContent).toBe("60");
  });

  it("クリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: onSelect の spy を渡した component
    const onSelect = vi.fn();
    const element = createEpisodeListItem(episode, onSelect);

    // When: 要素をクリックする
    element.click();

    // Then: onSelect が episodeId 付きで呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });
});
