import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeListState } from "../../view-models/episode-list-view-model.ts";
import { EpisodeList } from "./episode-list.tsx";

const baseUrl = "https://example.test";

describe("EpisodeList", () => {
  it("root に episode-list class を付ける", () => {
    // Given: loading state（children の有無に依存しない）
    const state: EpisodeListState = { status: "loading" };

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: list 容器の class が root にある（見た目は CSS 側の責務）
    expect(container.firstElementChild?.className).toBe("episode-list");
  });

  it("loading state の時、episode item を描画しない", () => {
    // Given: loading state
    const state: EpisodeListState = { status: "loading" };

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: item を含む要素が無い
    expect(container.querySelectorAll("[data-episode-title]")).toHaveLength(0);
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

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: title が episode の数だけ、内容もそのまま描画される
    const titles = Array.from(container.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["題1", "題2"]);
  });

  it("error state の時、episode item を描画しない", () => {
    // Given: error state
    const state: EpisodeListState = { status: "error" };

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: item を含む要素が無い
    expect(container.querySelectorAll("[data-episode-title]")).toHaveLength(0);
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

    // When: JSX として render しクリックする
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect }));
    container
      .querySelector("[data-episode-title]")
      ?.closest("article")
      ?.querySelector("button")
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

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: title・date は item 分の1つだけ（詳細側の重複描画が無い）。manuscript(topic)・player は1つ描画される
    expect(container.querySelectorAll("[data-episode-title]")).toHaveLength(2);
    expect(container.querySelectorAll("[data-episode-date]")).toHaveLength(2);
    expect(container.querySelectorAll("h1[data-episode-title]")).toHaveLength(0);
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(1);
    // why: audio の src 組み立て（buildRequestUrl）は episode-player.sociable_unit.test.ts で検証済み。
    //   ここでは EpisodeList の責務（baseUrl・audioRef を EpisodePlayer へ配置したか）だけを見る
    //   （Fault Isolation境界: testing-strategy/levels.md §3-1）
    expect(container.querySelectorAll("audio")).toHaveLength(1);

    // Then: 選択中 item の直後（次の兄弟）に詳細が展開される
    const items = Array.from(container.firstElementChild?.children ?? []);
    const selectedItemIndex = items.findIndex(
      (node) => node.querySelector("[data-episode-title]")?.textContent === "題1",
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

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: manuscript・player は無く、loading 表示がある
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(container.querySelectorAll("audio")).toHaveLength(0);
    expect(container.querySelector("[data-episode-detail-loading]")).not.toBeNull();
  });

  it("selectedEpisode が error の時、選択中 episode の直後に error 相当の要素を出す", () => {
    // Given: ep-1 を選択済み、詳細取得 error の state
    const state: EpisodeListState = {
      status: "success",
      episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 }],
      selectedEpisodeId: "ep-1",
      selectedEpisode: { status: "error" },
    };

    // When: JSX として render する
    const { container } = render(createElement(EpisodeList, { state, baseUrl, onSelect: vi.fn() }));

    // Then: manuscript・player は無く、error 表示がある
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(container.querySelectorAll("audio")).toHaveLength(0);
    expect(container.querySelector("[data-episode-detail-error]")).not.toBeNull();
  });

  it("props が同一参照のまま与えられた時、memo の浅い比較で再 render をスキップする", () => {
    // Given: EpisodeList を同一 element 参照で 2 回 render する
    const state: EpisodeListState = {
      status: "success",
      episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 }],
      selectedEpisodeId: null,
      selectedEpisode: null,
    };
    const onSelect = vi.fn();
    const element = createElement(EpisodeList, { state, baseUrl, onSelect });

    // When: 同じ props（同一参照）で再 render する
    const { container, rerender } = render(element);
    const firstHtml = container.innerHTML;
    rerender(element);

    // Then: memo コンポーネントである（React が浅い比較で skip 可能な形）。描画結果も不変
    expect((EpisodeList as { $$typeof?: symbol }).$$typeof).toBe(Symbol.for("react.memo"));
    expect(container.innerHTML).toBe(firstHtml);
  });
});
