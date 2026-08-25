import { describe, expect, it, vi } from "vitest";
import type { EpisodeListState } from "../view-models/episode-list-view-model.ts";
import { createEpisodeList } from "./episode-list.ts";

const baseUrl = "https://example.test";

describe("createEpisodeList", () => {
  it("loading state の時、episode item を描画しない", () => {
    // Given: loading state
    const state: EpisodeListState = { status: "loading" };

    // When: component を作る
    const element = createEpisodeList(state, baseUrl, vi.fn());

    // Then: item を含む要素が無い
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("success state の時、episode 毎に title を 1 つずつ描画する", () => {
    // Given: episode 2 件を持つ success state（選択なし）
    const state: EpisodeListState = {
      status: "success",
      episodes: [
        { episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 },
        { episodeId: "ep-2", date: "2026-08-18", title: "題2", durationSec: 90 },
      ],
      selectedEpisodeId: null,
      selectedEpisode: null,
    };

    // When: component を作る
    const element = createEpisodeList(state, baseUrl, vi.fn());

    // Then: title が episode の数だけ、内容もそのまま描画される
    const titles = Array.from(element.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["題1", "題2"]);
  });

  it("error state の時、episode item を描画しない", () => {
    // Given: error state
    const state: EpisodeListState = { status: "error" };

    // When: component を作る
    const element = createEpisodeList(state, baseUrl, vi.fn());

    // Then: item を含む要素が無い
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("item をクリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: episode 1 件を持つ success state
    const state: EpisodeListState = {
      status: "success",
      episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 }],
      selectedEpisodeId: null,
      selectedEpisode: null,
    };
    const onSelect = vi.fn();

    // When: component を作りクリックする
    const element = createEpisodeList(state, baseUrl, onSelect);
    element
      .querySelector("[data-episode-title]")
      ?.closest("article")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });

  it("selectedEpisode が success の時、選択中 episode の直後に manuscript・player を展開する。title・date は item に既にあるため header は展開しない", () => {
    // Given: ep-1 を選択済み、詳細取得 success の state
    const state: EpisodeListState = {
      status: "success",
      episodes: [
        { episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 },
        { episodeId: "ep-2", date: "2026-08-18", title: "題2", durationSec: 90 },
      ],
      selectedEpisodeId: "ep-1",
      selectedEpisode: {
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
      },
    };

    // When: component を作る
    const element = createEpisodeList(state, baseUrl, vi.fn());

    // Then: title・date は item 分の1つだけ（詳細側の重複描画が無い）。manuscript(topic)・player は1つ描画される
    expect(element.querySelectorAll("[data-episode-title]")).toHaveLength(2);
    expect(element.querySelectorAll("[data-episode-date]")).toHaveLength(2);
    expect(element.querySelectorAll("h1[data-episode-title]")).toHaveLength(0);
    expect(element.querySelectorAll("[data-topic-title]")).toHaveLength(1);
    expect(element.querySelectorAll("audio")).toHaveLength(1);
    expect(element.querySelector("audio")?.getAttribute("src")).toBe(
      "https://example.test/episodes/ep-1/audio",
    );

    // Then: 選択中 item の直後（次の兄弟）に詳細が展開される
    const items = Array.from(element.children);
    const selectedItemIndex = items.findIndex(
      (node) => node.querySelector("[data-episode-id]")?.textContent === "ep-1",
    );
    expect(items[selectedItemIndex + 1]?.querySelector("[data-topic-title]")).not.toBeNull();
  });

  it("selectedEpisode が loading の時、選択中 episode の直後に loading 相当の要素を出す", () => {
    // Given: ep-1 を選択済み、詳細取得 loading の state
    const state: EpisodeListState = {
      status: "success",
      episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 }],
      selectedEpisodeId: "ep-1",
      selectedEpisode: { status: "loading" },
    };

    // When: component を作る
    const element = createEpisodeList(state, baseUrl, vi.fn());

    // Then: manuscript・player は無く、loading 表示がある
    expect(element.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(element.querySelectorAll("audio")).toHaveLength(0);
    expect(element.querySelector("[data-episode-detail-loading]")).not.toBeNull();
  });

  it("selectedEpisode が error の時、選択中 episode の直後に error 相当の要素を出す", () => {
    // Given: ep-1 を選択済み、詳細取得 error の state
    const state: EpisodeListState = {
      status: "success",
      episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 }],
      selectedEpisodeId: "ep-1",
      selectedEpisode: { status: "error" },
    };

    // When: component を作る
    const element = createEpisodeList(state, baseUrl, vi.fn());

    // Then: manuscript・player は無く、error 表示がある
    expect(element.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(element.querySelectorAll("audio")).toHaveLength(0);
    expect(element.querySelector("[data-episode-detail-error]")).not.toBeNull();
  });
});
