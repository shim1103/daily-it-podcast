import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import { EpisodeDetailSuccess } from "./episode-detail-success.tsx";

const episode: EpisodeData = {
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
};

describe("EpisodeDetailSuccess", () => {
  it("seek bar と manuscript を episode-detail 内に描画する", () => {
    // Given: 詳細 success 状態
    const { container } = render(
      createElement(EpisodeDetailSuccess, {
        episode,
        playback: { kind: "playing", positionSec: 0, durationSec: 60 },
        onSeek: vi.fn(),
      }),
    );

    // Then: seek bar と原稿が detail 内にある
    expect(container.querySelector(".episode-detail .episode-seek-bar")).not.toBeNull();
    expect(container.querySelector("[data-manuscript-opening]")?.textContent).toBe("開始");
    expect(container.querySelector("[data-topic-title]")?.textContent).toBe("1. 小題");
  });

  it("topic seek ボタンで onSeek を呼ぶ", () => {
    // Given: onSeek spy
    const onSeek = vi.fn();
    const { container } = render(
      createElement(EpisodeDetailSuccess, {
        episode,
        playback: { kind: "stopped" },
        onSeek,
      }),
    );

    // When: seek ボタンを押す
    container
      .querySelector(".episode-topic__seek")
      ?.dispatchEvent(new Event("click", { bubbles: true }));

    // Then: onSeek が topic.startSec で呼ばれる
    expect(onSeek).toHaveBeenCalledWith(30);
  });
});
