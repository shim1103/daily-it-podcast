import { describe, expect, it } from "vitest";
import type { GetEpisodeResponse } from "../../../contracts/index.ts";
import { createEpisodeTopic } from "./episode-topic.ts";

type Topic = GetEpisodeResponse["body"]["topics"][number];

const topic: Topic = { title: "小題", preface: "前置き", detail: "詳細", startSec: 30 };

describe("createEpisodeTopic", () => {
  it("title・preface・detail をそのまま描画する", () => {
    // Given: topics[] の1要素
    // When: component を作る
    const element = createEpisodeTopic(topic);

    // Then: title・preface・detail がそのまま描画される
    expect(element.querySelector("[data-topic-title]")?.textContent).toBe("小題");
    expect(element.querySelector("[data-topic-preface]")?.textContent).toBe("前置き");
    expect(element.querySelector("[data-topic-detail]")?.textContent).toBe("詳細");
  });

  it("startSec は描画しない（contractに含めない）", () => {
    // Given: topics[] の1要素
    // When: component を作る
    const element = createEpisodeTopic(topic);

    // Then: startSec を示す要素が無い
    expect(element.querySelector("[data-topic-start-sec]")).toBeNull();
  });
});
