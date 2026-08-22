import { describe, expect, it } from "vitest";
import type { EpisodeListState } from "../view-models/episode-list-view-model.ts";
import { createEpisodeList } from "./episode-list.ts";

describe("createEpisodeList", () => {
  it("loading state の時、episode の title を描画しない", () => {
    // Given: loading state
    const state: EpisodeListState = { status: "loading" };

    // When: component を作る
    const element = createEpisodeList(state);

    // Then: title を含む要素が無い
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("success state の時、episode 毎に title を 1 つずつ描画する", () => {
    // Given: episode 2 件を持つ success state
    const state: EpisodeListState = {
      status: "success",
      episodes: [
        { episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 },
        { episodeId: "ep-2", date: "2026-08-18", title: "題2", durationSec: 90 },
      ],
    };

    // When: component を作る
    const element = createEpisodeList(state);

    // Then: title が episode の数だけ、内容もそのまま描画される
    const titles = Array.from(element.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["題1", "題2"]);
  });

  it("error state の時、episode の title を描画しない", () => {
    // Given: error state
    const state: EpisodeListState = { status: "error" };

    // When: component を作る
    const element = createEpisodeList(state);

    // Then: title を含む要素が無い
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("success state の時、title は詳細 page への hash link になる", () => {
    // Given: episode 1 件を持つ success state
    const state: EpisodeListState = {
      status: "success",
      episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 }],
    };

    // When: component を作る
    const element = createEpisodeList(state);

    // Then: title 要素の href が詳細 route を指す
    const link = element.querySelector("[data-episode-title]");
    expect(link?.getAttribute("href")).toBe("#/episodes/ep-1");
  });
});
