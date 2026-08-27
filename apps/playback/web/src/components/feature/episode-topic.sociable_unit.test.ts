import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import { EpisodeTopic } from "./episode-topic.tsx";

type Topic = EpisodeData["body"]["topics"][number];

const topic: Topic = { title: "小題", preface: "前置き", detail: "詳細", startSec: 30 };

describe("EpisodeTopic", () => {
  it("title・preface・detail をそのまま描画する", () => {
    // Given: topics[] の1要素
    // When: JSX として render する
    const { container } = render(createElement(EpisodeTopic, { topic }));

    // Then: title・preface・detail がそのまま描画される
    expect(container.querySelector("[data-topic-title]")?.textContent).toBe("小題");
    expect(container.querySelector("[data-topic-preface]")?.textContent).toBe("前置き");
    expect(container.querySelector("[data-topic-detail]")?.textContent).toBe("詳細");
  });

  it("startSec は描画しない（contractに含めない）", () => {
    // Given: topics[] の1要素
    // When: JSX として render する
    const { container } = render(createElement(EpisodeTopic, { topic }));

    // Then: startSec を示す要素が無い
    expect(container.querySelector("[data-topic-start-sec]")).toBeNull();
  });
});
