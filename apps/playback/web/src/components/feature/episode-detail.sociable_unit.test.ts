import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import type { EpisodeDetailState } from "../../view-models/episode-list-view-model.ts";
import { EpisodeDetail } from "./episode-detail.tsx";

describe("EpisodeDetail", () => {
  it("loading 状態で EpisodeDetailLoading を描画する", () => {
    // Given: loading 状態
    const detail: EpisodeDetailState = { status: "loading" };

    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeDetail, {
        detail,
        playback: { kind: "stopped" },
        onSeek: () => {},
      }),
    );

    // Then: loading 表示がある
    expect(container.querySelector("[data-episode-detail-loading]")).not.toBeNull();
  });

  it("error 状態で EpisodeDetailError を描画する", () => {
    // Given: error 状態
    const detail: EpisodeDetailState = { status: "error" };

    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeDetail, {
        detail,
        playback: { kind: "stopped" },
        onSeek: () => {},
      }),
    );

    // Then: error 表示がある
    expect(container.querySelector("[data-episode-detail-error]")).not.toBeNull();
  });

  it("success 状態で manuscript と seek bar を描画する", () => {
    // Given: success 状態
    const detail: EpisodeDetailState = {
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
    };

    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeDetail, {
        detail,
        playback: { kind: "playing", positionSec: 0, durationSec: 60 },
        onSeek: () => {},
      }),
    );

    // Then: success 内容が描画される
    expect(container.querySelector(".episode-seek-bar")).not.toBeNull();
    expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
  });
});
