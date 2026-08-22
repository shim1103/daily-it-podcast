import { describe, expect, it } from "vitest";
import type { EpisodeDetailState } from "../view-models/episode-detail-view-model.ts";
import { createEpisodeDetail } from "./episode-detail.ts";

describe("createEpisodeDetail", () => {
  it("loading state の時、title を描画しない", () => {
    // Given: loading state
    const state: EpisodeDetailState = { status: "loading" };

    // When: component を作る
    const element = createEpisodeDetail(state);

    // Then: title を含む要素が無い
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("success state の時、episode の title を描画する", () => {
    // Given: episode を持つ success state
    const state: EpisodeDetailState = {
      status: "success",
      episode: {
        episodeId: "ep-1",
        date: "2026-08-17",
        title: "題1",
        durationSec: 60,
        body: {
          opening: "開始",
          topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
          closing: "終了",
        },
        audioRef: "/episodes/ep-1/audio",
      },
    };

    // When: component を作る
    const element = createEpisodeDetail(state);

    // Then: title だけが描画される
    const titles = Array.from(element.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["題1"]);
  });

  it("error state の時、title を描画しない", () => {
    // Given: error state
    const state: EpisodeDetailState = { status: "error" };

    // When: component を作る
    const element = createEpisodeDetail(state);

    // Then: title を含む要素が無い
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });
});
