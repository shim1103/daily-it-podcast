import { render } from "@testing-library/react";
import { createElement, createRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { EpisodeListEntry, type EpisodeListEntryProps } from "./episode-list-entry.tsx";

const listEpisode: EpisodeListEntryProps["episode"] = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  topics: [{ title: "小題A" }, { title: "小題B" }],
  audioRef: "/episodes/ep-1/audio",
};

function renderEpisodeListEntry(overrides: Partial<EpisodeListEntryProps> = {}) {
  return render(
    createElement(EpisodeListEntry, {
      episode: listEpisode,
      episodeCount: 1,
      episodeIndex: 0,
      selection: { kind: "closed" },
      playback: { kind: "stopped" },
      onSelect: vi.fn(),
      onPlay: vi.fn(),
      onSeek: vi.fn(),
      ...overrides,
    }),
  );
}

describe("EpisodeListEntry", () => {
  it("selection closed の時、row のみ描画する", () => {
    // Given: 選択なし
    const { container } = renderEpisodeListEntry();

    // Then: row があり selected modifier が無い
    expect(container.querySelector("article.episode-row")).not.toBeNull();
    expect(container.querySelector(".episode-list-entry--selected")).toBeNull();
    expect(container.querySelector(".episode-detail")).toBeNull();
  });

  it("selection open かつ detail success の時、紫枠 modifier と detail を描画する", () => {
    // Given: ep-1 を選択済み、詳細取得 success
    const { container } = renderEpisodeListEntry({
      selection: {
        kind: "open",
        detail: {
          status: "success",
          episode: {
            episodeId: "ep-1",
            date: "2026-08-17",
            title: "題1",
            durationSec: 60,
            body: {
              opening: "開始",
              topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 30 }],
              closing: "終了",
            },
            audioRef: "/episodes/ep-1/audio",
          },
        },
      },
      playback: { kind: "playing", positionSec: 10, durationSec: 60 },
    });

    // Then: 紫枠・seek bar・原稿が entry 内にある
    expect(container.querySelector(".episode-list-entry--selected")).not.toBeNull();
    expect(container.querySelector(".episode-list-entry--selected .episode-seek-bar")).not.toBeNull();
    expect(container.querySelector(".episode-list-entry--selected [data-topic-title]")?.textContent).toBe(
      "1. 小題",
    );
    expect(container.querySelector(".episode-list-entry--selected .episode-detail")).not.toBeNull();
  });

  it("selection open かつ detail loading の時、loading 相当の要素を出す", () => {
    // Given: ep-1 を選択済み、詳細取得 loading
    const { container } = renderEpisodeListEntry({
      selection: { kind: "open", detail: { status: "loading" } },
    });

    // Then: manuscript・seek bar は無く、loading 表示がある
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(container.querySelectorAll(".episode-seek-bar")).toHaveLength(0);
    expect(
      container.querySelector(
        ".episode-list-entry--selected .episode-detail[data-episode-detail-loading]",
      ),
    ).not.toBeNull();
  });

  it("selection open かつ detail error の時、error 相当の要素を出す", () => {
    // Given: ep-1 を選択済み、詳細取得 error
    const { container } = renderEpisodeListEntry({
      selection: { kind: "open", detail: { status: "error" } },
    });

    // Then: manuscript・seek bar は無く、error 表示がある
    expect(container.querySelectorAll("[data-topic-title]")).toHaveLength(0);
    expect(container.querySelectorAll(".episode-seek-bar")).toHaveLength(0);
    expect(
      container.querySelector(".episode-list-entry--selected .episode-detail[data-episode-detail-error]"),
    ).not.toBeNull();
  });

  it("playback stopped の時、audio を描画しない", () => {
    // Given: 停止中
    const audioElementRef = createRef<HTMLAudioElement | null>();
    const { container } = renderEpisodeListEntry({
      playback: { kind: "stopped" },
      audioElementRef,
      resolvedSrc: "https://example.test/episodes/ep-1/audio",
    });

    // Then: entry 内に audio は無い
    expect(container.querySelector(".episode-list-entry audio")).toBeNull();
  });

  it("playback playing の時、entry 内の先頭に audio を描画する", () => {
    // Given: 再生中と audio 配線
    const audioElementRef = createRef<HTMLAudioElement | null>();
    const { container } = renderEpisodeListEntry({
      playback: { kind: "playing", positionSec: 0, durationSec: 60 },
      audioElementRef,
      resolvedSrc: "https://example.test/episodes/ep-1/audio",
    });

    // Then: entry 内に audio があり row より前にある
    const entry = container.querySelector(".episode-list-entry");
    const audio = entry?.querySelector("audio");
    const row = entry?.querySelector("article.episode-row");
    expect(audio).not.toBeNull();
    expect(audio?.getAttribute("src")).toBe("https://example.test/episodes/ep-1/audio");
    expect(entry?.firstElementChild?.className).toBe("episode-audio");
    expect(row).not.toBeNull();
    expect(audioElementRef.current).toBe(audio);
  });

  it("play pill を click すると onPlay が episodeId と audioRef 付きで呼ばれる", () => {
    // Given: onPlay の spy
    const onPlay = vi.fn();
    const { container } = renderEpisodeListEntry({ onPlay });

    // When: play pill を click する
    container
      .querySelector(".episode-play-button")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onPlay が呼ばれる
    expect(onPlay).toHaveBeenCalledWith("ep-1", "/episodes/ep-1/audio");
  });
});
