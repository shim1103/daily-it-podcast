import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it, vi } from "vitest";
import type {
  EpisodeListItemData,
  EpisodeListState,
} from "../../view-models/episode-list-view-model.legacy.ts";
import { EpisodeList, type EpisodeListProps } from "./episode-list.legacy.tsx";

const baseUrl = "https://example.test";

const minimalBody = {
  opening: "開始",
  topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
  closing: "終了",
};

function makeEpisode(
  episodeId: string,
  title: string,
  date: string,
  durationSec: number,
): EpisodeListItemData {
  return {
    episodeId,
    date,
    title,
    durationSec,
    body: minimalBody,
    audioRef: `/episodes/${episodeId}/audio`,
  };
}

function renderEpisodeList(props: Pick<EpisodeListProps, "state"> & Partial<EpisodeListProps>) {
  return render(
    createElement(EpisodeList, {
      baseUrl,
      onSelect: vi.fn(),
      audioElementRef: createRef<HTMLAudioElement | null>(),
      onSeek: vi.fn(),
      ...props,
    }),
  );
}

describe("EpisodeList", () => {
  it("root に episode-list class を付ける", () => {
    // Given: loading state（children の有無に依存しない）
    const state: EpisodeListState = { status: "loading" };

    // When: JSX として render する
    const { container } = renderEpisodeList({ state });

    // Then: list 容器の class が root にある（見た目は CSS 側の責務）
    expect(container.firstElementChild?.className).toBe("episode-list");
  });

  it("loading state の時、episode item を描画しない", () => {
    // Given: loading state
    const state: EpisodeListState = { status: "loading" };

    // When: JSX として render する
    const { container } = renderEpisodeList({ state });

    // Then: item を含む要素が無い
    expect(container.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("success state の時、episode 毎に title を 1 つずつ描画する", () => {
    // Given: episode 2 件を持つ success state（選択なし）
    const state: EpisodeListState = {
      status: "success",
      episodes: [
        makeEpisode("ep-1", "題1", "2026-08-17", 60),
        makeEpisode("ep-2", "題2", "2026-08-18", 90),
      ],
      selectedEpisodeId: null,
      selectedEpisode: null,
    };

    // When: JSX として render する
    const { container } = renderEpisodeList({ state });

    // Then: title が episode の数だけ、内容もそのまま描画される
    const titles = Array.from(container.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["2.　題1", "1.　題2"]);
  });

  it("error state の時、episode item を描画しない", () => {
    // Given: error state
    const state: EpisodeListState = { status: "error" };

    // When: JSX として render する
    const { container } = renderEpisodeList({ state });

    // Then: item を含む要素が無い
    expect(container.querySelectorAll("[data-episode-title]")).toHaveLength(0);
  });

  it("item をクリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: episode 1 件を持つ success state
    const state: EpisodeListState = {
      status: "success",
      episodes: [makeEpisode("ep-1", "題1", "2026-08-17", 60)],
      selectedEpisodeId: null,
      selectedEpisode: null,
    };
    const onSelect = vi.fn();
    const { container } = renderEpisodeList({ state, onSelect });
    container
      .querySelector("[data-episode-title]")
      ?.closest("article")
      ?.querySelector("button")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });

  it("selectedEpisode が success の時、選択中 episode だけを紫枠グループで描画する", () => {
    // Given: ep-1 を選択済み、詳細 success の state
    const ep1 = makeEpisode("ep-1", "題1", "2026-08-17", 60);
    const state: EpisodeListState = {
      status: "success",
      episodes: [ep1, makeEpisode("ep-2", "題2", "2026-08-18", 90)],
      selectedEpisodeId: "ep-1",
      selectedEpisode: {
        status: "success",
        episode: ep1,
      },
    };

    // When: JSX として render する
    const { container } = renderEpisodeList({ state });

    // Then: 選択中 episode だけが描画される
    expect(container.querySelectorAll("[data-episode-title]")).toHaveLength(1);
    expect(container.querySelectorAll("[data-episode-date]")).toHaveLength(1);
    expect(container.querySelectorAll("h1[data-episode-title]")).toHaveLength(0);
    expect(container.querySelector(".episode-selected-group")).not.toBeNull();
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(1);
    expect(container.querySelector("[data-topic-title]")?.textContent).toBe("1. 小題");
    expect(container.querySelectorAll("audio")).toHaveLength(1);
    expect(container.querySelector(".episode-selected-group [data-topic-title]")).not.toBeNull();
    expect(container.querySelector(".episode-selected-group .episode-detail")).not.toBeNull();
  });

  it("selectedEpisode が error の時、選択中 episode の直後に error 相当の要素を出す", () => {
    // Given: ep-1 を選択済み、lookup 失敗 error の state
    const state: EpisodeListState = {
      status: "success",
      episodes: [makeEpisode("ep-1", "題1", "2026-08-17", 60)],
      selectedEpisodeId: "ep-1",
      selectedEpisode: { status: "error" },
    };

    // When: JSX として render する
    const { container } = renderEpisodeList({ state });

    // Then: manuscript・player は無く、error 表示がある。選択グループ内
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(container.querySelectorAll("audio")).toHaveLength(0);
    expect(
      container.querySelector(".episode-selected-group .episode-detail[data-episode-detail-error]"),
    ).not.toBeNull();
  });

  it("props が同一参照のまま与えられた時、memo の浅い比較で再 render をスキップする", () => {
    // Given: EpisodeList を同一 element 参照で 2 回 render する
    const state: EpisodeListState = {
      status: "success",
      episodes: [makeEpisode("ep-1", "題1", "2026-08-17", 60)],
      selectedEpisodeId: null,
      selectedEpisode: null,
    };
    const onSelect = vi.fn();
    const element = createElement(EpisodeList, {
      state,
      baseUrl,
      onSelect,
      audioElementRef: createRef<HTMLAudioElement | null>(),
      onSeek: vi.fn(),
    });

    // When: 同じ props（同一参照）で再 render する
    const { container, rerender } = render(element);
    const firstHtml = container.innerHTML;
    rerender(element);

    // Then: memo コンポーネントである（React が浅い比較で skip 可能な形）。描画結果も不変
    expect((EpisodeList as { $$typeof?: symbol }).$$typeof).toBe(Symbol.for("react.memo"));
    expect(container.innerHTML).toBe(firstHtml);
  });
});
