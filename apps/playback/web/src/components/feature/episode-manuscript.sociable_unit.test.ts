import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeManuscript } from "./episode-manuscript.tsx";

type Body = EpisodeData["body"];

const body: Body = {
  opening: "開始文",
  topics: [
    { title: "小題1", preface: "前置き1", detail: "詳細1", startSec: 0 },
    { title: "小題2", preface: "前置き2", detail: "詳細2", startSec: 30 },
  ],
  closing: "終了文",
};

const onSeek = vi.fn();

describe("EpisodeManuscript", () => {
  it("root に episode-manuscript class を付ける", () => {
    // Given: EpisodeItem の body
    // When: JSX として render する
    const { container } = render(createElement(EpisodeManuscript, { body, onSeek }));

    // Then: manuscript 容器の class が root にある（見た目は CSS 側の責務）
    expect(container.firstElementChild?.className).toBe("episode-manuscript");
  });

  it("opening・closing をそのまま描画する", () => {
    // Given: EpisodeItem の body
    // When: JSX として render する
    const { container } = render(createElement(EpisodeManuscript, { body, onSeek }));

    // Then: opening・closing がそのまま描画される
    expect(container.querySelector("[data-manuscript-opening]")?.textContent).toBe("開始文");
    expect(container.querySelector("[data-manuscript-closing]")?.textContent).toBe("終了文");
  });

  it("topics[] の数だけ episode-topic を並べる", () => {
    // Given: topics[] を2件持つ body
    // When: JSX として render する
    const { container } = render(createElement(EpisodeManuscript, { body, onSeek }));

    // Then: topic titleが2件、順番通りに描画される
    const titles = Array.from(container.querySelectorAll("[data-topic-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["1. 小題1", "2. 小題2"]);
  });
});
