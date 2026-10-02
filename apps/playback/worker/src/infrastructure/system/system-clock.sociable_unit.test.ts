import { afterEach, describe, expect, it, vi } from "vitest";
import { SystemClock } from "./system-clock.ts";

/**
 * scope: Sociable Unit
 * real: SystemClock
 * double: vi の fake timer（システム時刻のみ固定）
 */
describe("SystemClock", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("now_returns_current_system_time_as_date", () => {
    // Given: システム時刻を固定した fake timer
    const fixed = new Date("2026-10-01T12:34:56.789Z");
    vi.useFakeTimers();
    vi.setSystemTime(fixed);

    // When: 現在時刻を取得する
    const got = new SystemClock().now();

    // Then: 固定したシステム時刻と同じ instant の Date が返る
    expect(got).toBeInstanceOf(Date);
    expect(got.getTime()).toBe(fixed.getTime());
  });
});
