import { describe, expect, it } from "vitest";
import { PROGRESS_COMPLETE_ZONE_SEC } from "./progress.ts";

describe("progress domain constants", () => {
  it("完走ゾーンは 3 秒である", () => {
    expect(PROGRESS_COMPLETE_ZONE_SEC).toBe(3);
  });
});
