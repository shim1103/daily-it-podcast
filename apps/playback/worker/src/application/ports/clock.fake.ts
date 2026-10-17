import type { Clock } from "./clock.ts";

/**
 * 固定時刻を返す Clock（Application の test 用 Fake）。
 * 本番 adapter は置かない。SystemClock が本番の時刻源である。
 *
 * @ensure `now()` は常に構築時に渡した instant と同じ時刻の新しい `Date` を返す
 */
export class FakeClock implements Clock {
  constructor(private readonly fixedAt: Date) {}

  now(): Date {
    return new Date(this.fixedAt.getTime());
  }
}
