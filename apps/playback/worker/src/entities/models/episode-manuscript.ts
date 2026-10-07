/** 原稿の話題 1 件。`startSec` は音声内の開始位置（秒）。 */
export type EpisodeManuscriptTopic = {
  readonly title: string;
  readonly preface: string;
  readonly detail: string;
  readonly startSec: number;
};

/** 原稿の冒頭・結びの 1 区間。 */
export type EpisodeManuscriptSegment = {
  readonly text: string;
  readonly startSec: number;
};

/**
 * 検証済みの原稿 1 件。repo 根 `contracts/manuscript.schema.json` に適合し、`episodeId` が取得元の stem と一致する。
 * 音声の参照（HTTP の path）と進捗は持たない。それらは境界の写しと別の取得で足す。
 */
export type EpisodeManuscript = {
  readonly episodeId: string;
  readonly date: string;
  readonly title: string;
  readonly durationSec: number;
  readonly body: {
    readonly opening: EpisodeManuscriptSegment;
    readonly topics: readonly EpisodeManuscriptTopic[];
    readonly ending: EpisodeManuscriptSegment;
  };
};
