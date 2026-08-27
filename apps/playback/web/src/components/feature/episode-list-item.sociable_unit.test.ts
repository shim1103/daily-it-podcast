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
};

describe("EpisodeListItem", () => {
  it("episodeId・date・title・durationSec をそのまま描画する", () => {
    // Given: EpisodeListItem 1件
    // When: JSX として render する
    const { container } = render(createElement(EpisodeListItem, { episode, onSelect: vi.fn() }));

    // Then: 各 field がそのまま描画される
    expect(container.querySelector("[data-episode-id]")?.textContent).toBe("ep-1");
    expect(container.querySelector("[data-episode-date]")?.textContent).toBe("2026-08-17");
    expect(container.querySelector("[data-episode-title]")?.textContent).toBe("題1");
    expect(container.querySelector("[data-episode-duration-sec]")?.textContent).toBe("60");
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
