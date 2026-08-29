import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { EpisodeList, type EpisodeListProps } from "./episode-list.tsx";

const listEpisode = (id: string, title: string, durationSec: number) => ({
  episodeId: id,
  date: "2026-08-17",
  title,
  durationSec,
  topics: [] as { title: string }[],
  audioRef: `/episodes/${id}/audio`,
});

function renderEpisodeList(
  props: Partial<EpisodeListProps> & Pick<EpisodeListProps, "episodes">,
) {
  return render(
    createElement(EpisodeList, {
      selection: { kind: "none" },
      playback: { kind: "idle" },
      onSelect: vi.fn(),
      onPlay: vi.fn(),
      onSeek: vi.fn(),
      audioElementRef: createRef<HTMLAudioElement | null>(),
      resolvedSrc: undefined,
      ...props,
    }),
  );
}

describe("EpisodeList", () => {
  it("root に episode-list class を付ける", () => {
    // Given: episode 0 件
    // When: JSX として render する
    const { container } = renderEpisodeList({ episodes: [] });

    // Then: list 容器の class が root にある
    expect(container.firstElementChild?.className).toBe("episode-list");
  });

  it("episode 毎に title を 1 つずつ描画する", () => {
    // Given: episode 2 件（選択なし）
    const episodes = [listEpisode("ep-1", "題1", 60), listEpisode("ep-2", "題2", 90)];

    // When: JSX として render する
    const { container } = renderEpisodeList({ episodes });

    // Then: title が episode の数だけ、内容もそのまま描画される
    const titles = Array.from(container.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["2.　題1", "1.　題2"]);
  });

  it("item をクリックすると onSelect が episodeId 付きで呼ばれる", () => {
    // Given: episode 1 件
    const episodes = [listEpisode("ep-1", "題1", 60)];
    const onSelect = vi.fn();
    const { container } = renderEpisodeList({ episodes, onSelect });
    container
      .querySelector(".episode-row__hit")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSelect が呼ばれる
    expect(onSelect).toHaveBeenCalledWith("ep-1");
  });

  it("props が同一参照のまま与えられた時、memo の浅い比較で再 render をスキップする", () => {
    // Given: EpisodeList を同一 element 参照で 2 回 render する
    const episodes = [listEpisode("ep-1", "題1", 60)];
    const onSelect = vi.fn();
    const element = createElement(EpisodeList, {
      episodes,
      selection: { kind: "none" },
      playback: { kind: "idle" },
      onSelect,
      onPlay: vi.fn(),
      onSeek: vi.fn(),
      audioElementRef: createRef<HTMLAudioElement | null>(),
      resolvedSrc: undefined,
    });

    // When: 同じ props（同一参照）で再 render する
    const { container, rerender } = render(element);
    const firstHtml = container.innerHTML;
    rerender(element);

    // Then: memo コンポーネントである。描画結果も不変
    expect((EpisodeList as { $$typeof?: symbol }).$$typeof).toBe(Symbol.for("react.memo"));
    expect(container.innerHTML).toBe(firstHtml);
  });

  it("再生中は playing entry だけが audio を持ち、list root 直下には audio が無い", () => {
    // Given: ep-1 再生中と audio 配線
    const episodes = [listEpisode("ep-1", "題1", 60), listEpisode("ep-2", "題2", 90)];
    const audioElementRef = createRef<HTMLAudioElement | null>();
    const { container } = renderEpisodeList({
      episodes,
      playback: { kind: "playing", episodeId: "ep-1", positionSec: 0, durationSec: 60 },
      audioElementRef,
      resolvedSrc: "https://example.test/episodes/ep-1/audio",
    });

    // Then: list root 直下に audio は無く、ep-1 entry 内だけにある
    const list = container.querySelector(".episode-list");
    expect(list?.querySelector(":scope > .episode-audio")).toBeNull();
    expect(list?.querySelector(":scope > audio")).toBeNull();
    const entries = container.querySelectorAll(".episode-list-entry");
    expect(entries[0]?.querySelector("audio")).not.toBeNull();
    expect(entries[1]?.querySelector("audio")).toBeNull();
  });

  it("list 自体は audio を持たず、entry 数と row 数が一致する", () => {
    // Given: episode 2 件（選択なし）
    const episodes = [listEpisode("ep-1", "題1", 60), listEpisode("ep-2", "題2", 90)];

    // When: JSX として render する
    const { container } = renderEpisodeList({ episodes });

    // Then: 全 entry が見え、list 内に audio は無い
    expect(container.querySelectorAll(".episode-list-entry")).toHaveLength(2);
    expect(container.querySelectorAll("article.episode-row")).toHaveLength(2);
    expect(container.querySelectorAll("audio")).toHaveLength(0);
  });

  it("selection open の時も全 item が描画される", () => {
    // Given: episode 2 件、ep-1 を選択中
    const episodes = [listEpisode("ep-1", "題1", 60), listEpisode("ep-2", "題2", 90)];

    // When: JSX として render する
    const { container } = renderEpisodeList({
      episodes,
      selection: { kind: "open", episodeId: "ep-1", detail: { status: "loading" } },
    });

    // Then: 両方の title が見える
    const titles = Array.from(container.querySelectorAll("[data-episode-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["2.　題1", "1.　題2"]);
  });

  it("play pill を click すると onPlay が episodeId と audioRef 付きで呼ばれる", () => {
    // Given: episode 2 件
    const episodes = [listEpisode("ep-1", "題1", 60), listEpisode("ep-2", "題2", 90)];
    const onPlay = vi.fn();
    const { container } = renderEpisodeList({ episodes, onPlay });

    // When: ep-2 の play pill を click する
    const playButtons = container.querySelectorAll(".episode-play-button");
    playButtons[1]?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onPlay が ep-2 向けに呼ばれる
    expect(onPlay).toHaveBeenCalledWith("ep-2", "/episodes/ep-2/audio");
  });
});
