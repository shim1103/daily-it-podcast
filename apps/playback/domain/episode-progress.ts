import type { UtcInstant } from "./utc-instant.ts";

/**
 * 再生進捗の本体（1 episode = 1 行）。未 play は行なしで表し、この型は持たない。
 * `firstCompletedAt` だけが未完走で null になる。
 */
export type EpisodeProgress = {
  readonly positionSec: number;
  readonly firstPlayedAt: UtcInstant;
  readonly firstCompletedAt: UtcInstant | null;
  readonly lastPlayedAt: UtcInstant;
};
