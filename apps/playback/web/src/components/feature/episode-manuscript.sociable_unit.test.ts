import { fireEvent, render } from "@testing-library/react";
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

const DURATION_SEC = 1000;

function renderManuscript(onSeek = vi.fn()) {
  return render(createElement(EpisodeManuscript, { body, durationSec: DURATION_SEC, onSeek }));
}

describe("EpisodeManuscript", () => {
  it("root に episode-manuscript class を付ける", () => {
    // Given: EpisodeItem の body
    // When: JSX として render する
    const { container } = renderManuscript();

    // Then: manuscript 容器の class が root にある（見た目は CSS 側の責務）
    expect(container.firstElementChild?.className).toBe("episode-manuscript");
  });

  it("導入（opening）を 00:00 の seek bar 付きで描画する", () => {
    // Given: body（durationSec 1000）
    const { container } = renderManuscript();

    // Then: 「導入」見出し・00:00 の bar・opening 本文
    const bar = container.querySelector("[data-manuscript-opening-start-sec]");
    expect(bar?.textContent).toBe("00:00");
    expect(container.querySelector("[data-manuscript-opening]")?.textContent).toBe("開始文");
    expect(container.textContent).toContain("導入");
  });

  it("まとめ（closing）を総尺（durationSec）の seek bar 付きで描画する", () => {
    // Given: body（durationSec 1000 → 16:40）
    const { container } = renderManuscript();

    // Then: 「まとめ」見出し・16:40 の bar・closing 本文
    const bar = container.querySelector("[data-manuscript-closing-start-sec]");
    expect(bar?.textContent).toBe("16:40");
    expect(container.querySelector("[data-manuscript-closing]")?.textContent).toBe("終了文");
    expect(container.textContent).toContain("まとめ");
  });

  it("導入の bar を押すと onSeek(0)、まとめの bar を押すと onSeek(durationSec)", () => {
    // Given: onSeek spy
    const onSeek = vi.fn();
    const { container } = renderManuscript(onSeek);

    // When: 導入 bar → まとめ bar の順に押す
    fireEvent.click(container.querySelector("[data-manuscript-opening-start-sec]") as Element);
    fireEvent.click(container.querySelector("[data-manuscript-closing-start-sec]") as Element);

    // Then: 0 と総尺が渡る
    expect(onSeek).toHaveBeenNthCalledWith(1, 0);
    expect(onSeek).toHaveBeenNthCalledWith(2, DURATION_SEC);
  });

  it("topics[] の数だけ episode-topic を並べる（導入・まとめの間）", () => {
    // Given: topics[] を2件持つ body
    const { container } = renderManuscript();

    // Then: topic title が2件、順番通りに描画される
    const titles = Array.from(container.querySelectorAll("[data-topic-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["1. 小題1", "2. 小題2"]);
  });
});
