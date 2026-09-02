import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeEntry } from "./episode-entry.tsx";

const body: EpisodeData["body"] = {
  opening: "開始",
  topics: [
    { title: "小題A", preface: "前A", detail: "詳A", startSec: 0 },
    { title: "小題B", preface: "前B", detail: "詳B", startSec: 30 },
  ],
  closing: "終了",
};

describe("EpisodeEntry", () => {
  it("root に episode-entry class を付けて manuscript を描画する", () => {
    // Given: body と onSeek
    // When: JSX として render する
    const { container } = render(createElement(EpisodeEntry, { body, onSeek: vi.fn() }));

    // Then: episode-entry class と opening が出る
    expect(container.firstElementChild?.className).toBe("episode-entry");
    expect(container.querySelector("[data-manuscript-opening]")?.textContent).toBe("開始");
  });
});
