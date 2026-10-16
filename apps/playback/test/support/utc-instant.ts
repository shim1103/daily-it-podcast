import { normalizeUtcInstant, type UtcInstant } from "../../domain/index.ts";

/**
 * test の時刻 literal を `UtcInstant` にする。
 *
 * @require input は `normalizeUtcInstant` が受理する、offset 付き ISO-8601 で範囲内
 * @ensure input と同じ instant の UTC 固定幅を返す。受理されない input は throw する（test の書き間違いを即座に落とす）
 */
export function utc(input: string): UtcInstant {
  const instant = normalizeUtcInstant(input);
  if (instant === undefined) {
    throw new Error(`UtcInstant にできない時刻 literal: ${input}`);
  }
  return instant;
}
