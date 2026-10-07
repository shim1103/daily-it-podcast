/**
 * 再生進捗の本体（1 episode = 1 行）。未 play は行なしで表し、この型は持たない。
 * 時刻は UTC 固定幅の ISO 表記で、文字列の大小が時系列と一致する（契約境界で正規化済み）。
 * `firstCompletedAt` だけが未完走で null になる。
 */
export type EpisodeProgress = {
  readonly positionSec: number;
  readonly firstPlayedAt: string;
  readonly firstCompletedAt: string | null;
  readonly lastPlayedAt: string;
};
