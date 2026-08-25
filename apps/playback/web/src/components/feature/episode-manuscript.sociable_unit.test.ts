import { describe, expect, it } from "vitest";
import type { GetEpisodeResponse } from "../../../contracts/index.ts";
import { createEpisodeManuscript } from "./episode-manuscript.ts";

const body: GetEpisodeResponse["body"] = {
  opening: "開始文",
  topics: [
    { title: "小題1", preface: "前置き1", detail: "詳細1", startSec: 0 },
    { title: "小題2", preface: "前置き2", detail: "詳細2", startSec: 30 },
  ],
  closing: "終了文",
};

describe("createEpisodeManuscript", () => {
  it("opening・closing をそのまま描画する", () => {
    // Given: GetEpisodeResponse の body
    // When: component を作る
    const element = createEpisodeManuscript(body);

    // Then: opening・closing がそのまま描画される
    expect(element.querySelector("[data-manuscript-opening]")?.textContent).toBe("開始文");
    expect(element.querySelector("[data-manuscript-closing]")?.textContent).toBe("終了文");
  });

  it("topics[] の数だけ episode-topic を並べる", () => {
    // Given: topics[] を2件持つ body
    // When: component を作る
    const element = createEpisodeManuscript(body);

    // Then: topic titleが2件、順番通りに描画される
    const titles = Array.from(element.querySelectorAll("[data-topic-title]")).map(
      (node) => node.textContent,
    );
    expect(titles).toEqual(["小題1", "小題2"]);
  });
});
