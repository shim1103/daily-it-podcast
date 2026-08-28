import { describe, expect, it } from "vitest";
import { formatDurationMmSs } from "./format-duration-mm-ss.ts";

describe("formatDurationMmSs", () => {
  it("60 秒を 01:00 にする", () => {
    // Given: 60 秒
    const durationSec = 60;

    // When: mm:ss へ整形する
    const got = formatDurationMmSs(durationSec);

    // Then: ゼロ埋めの mm:ss
    expect(got).toBe("01:00");
  });

  it("0 秒を 00:00 にする", () => {
    // Given: 0 秒
    const durationSec = 0;

    // When: mm:ss へ整形する
    const got = formatDurationMmSs(durationSec);

    // Then: ゼロ埋めの mm:ss
    expect(got).toBe("00:00");
  });

  it("125 秒を 02:05 にする", () => {
    // Given: 2 分 5 秒
    const durationSec = 125;

    // When: mm:ss へ整形する
    const got = formatDurationMmSs(durationSec);

    // Then: ゼロ埋めの mm:ss（単位文字なし）
    expect(got).toBe("02:05");
  });

  it("長い尺でも mm:ss のまま出す", () => {
    // Given: 1 時間超
    const durationSec = 3723;

    // When: mm:ss へ整形する
    const got = formatDurationMmSs(durationSec);

    // Then: 分は 2 桁以上、秒は 2 桁
    expect(got).toBe("62:03");
  });
});
