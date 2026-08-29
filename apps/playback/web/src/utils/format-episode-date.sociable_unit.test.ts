import { describe, expect, it } from "vitest";
import { formatEpisodeDate } from "./format-episode-date.ts";

describe("formatEpisodeDate", () => {
  it("YYYY-MM-DD を YYYY/MM/DD に変換する", () => {
    // Given: wire 形式の日付
    const wireDate = "2026-08-17";

    // When: 表示用へ整形する
    const got = formatEpisodeDate(wireDate);

    // Then: スラッシュ区切りになる
    expect(got).toBe("2026/08/17");
  });

  it("別の日付でも同じ規則でスラッシュ区切りにする", () => {
    // Given: 別の wire 日付
    const wireDate = "2025-01-02";

    // When: 表示用へ整形する
    const got = formatEpisodeDate(wireDate);

    // Then: スラッシュ区切りになる
    expect(got).toBe("2025/01/02");
  });
});
