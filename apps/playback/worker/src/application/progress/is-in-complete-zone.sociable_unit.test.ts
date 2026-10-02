import { describe, expect, it } from "vitest";
import { PROGRESS_COMPLETE_ZONE_SEC } from "../../entities/constants/progress.ts";
import { isInCompleteZone } from "./is-in-complete-zone.ts";

describe("isInCompleteZone", () => {
  it("positionSec が durationSec - ゾーン幅以上の時、true を返す", () => {
    // Given: duration 60・ゾーン幅 3 なら境界は 57
    const durationSec = 60;
    const positionSec = durationSec - PROGRESS_COMPLETE_ZONE_SEC;

    // When: ゾーン判定する
    const got = isInCompleteZone(positionSec, durationSec);

    // Then: 境界ちょうどは完走ゾーン内
    expect(got).toBe(true);
  });

  it("positionSec が durationSec - ゾーン幅未満の時、false を返す", () => {
    // Given: 境界の直下
    const durationSec = 60;
    const positionSec = durationSec - PROGRESS_COMPLETE_ZONE_SEC - 0.001;

    // When: ゾーン判定する
    const got = isInCompleteZone(positionSec, durationSec);

    // Then: 境界未満はゾーン外
    expect(got).toBe(false);
  });

  it("positionSec が durationSec と同じ時、true を返す", () => {
    // Given: 末尾ちょうど
    // When: ゾーン判定する
    const got = isInCompleteZone(60, 60);

    // Then: 末尾も完走ゾーン内
    expect(got).toBe(true);
  });

  it("durationSec がゾーン幅未満でも同じ不等式で判定する", () => {
    // Given: 短尺 durationSec=2（幅 3 未満）・positionSec=0
    // When: ゾーン判定する
    const got = isInCompleteZone(0, 2);

    // Then: 0 >= 2-3 なのでゾーン内（別ルールなし）
    expect(got).toBe(true);
    expect(PROGRESS_COMPLETE_ZONE_SEC).toBe(3);
  });
});
