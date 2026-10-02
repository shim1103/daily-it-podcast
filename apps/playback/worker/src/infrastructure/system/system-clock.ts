import type { Clock } from "../../application/ports/clock.ts";

/**
 * 実行環境のシステム時計を返す `Clock`。in-memory mode・r2 mode のどちらでも本番の時刻源になる。
 *
 * @ensure `now()` は呼出しごとに、その時点の現在時刻を表す新しい `Date` を返す
 */
export class SystemClock implements Clock {
  now(): Date {
    return new Date();
  }
}
