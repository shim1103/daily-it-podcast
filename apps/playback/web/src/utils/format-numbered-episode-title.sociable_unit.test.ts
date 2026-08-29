import { describe, expect, it } from "vitest";
import { formatNumberedEpisodeTitle } from "./format-numbered-episode-title.ts";

describe("formatNumberedEpisodeTitle", () => {
  it("最古（末尾 index）を 1、先頭を episodeCount にする", () => {
    // Given: 3 件中 2 番目（0 始まり index 1）
    // When: 整形する
    const got = formatNumberedEpisodeTitle(3, 1, "題");

    // Then: 2.　題
    expect(got).toBe("2.　題");
  });

  it("1 件だけの時は 1.　title になる", () => {
    // Given: 1 件・index 0
    const got = formatNumberedEpisodeTitle(1, 0, "単独");

    // Then: 1.　単独
    expect(got).toBe("1.　単独");
  });
});
