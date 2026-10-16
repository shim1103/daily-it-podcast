import type { EpisodeId } from "./episode-id.ts";
import type { EpisodeProgress } from "./episode-progress.ts";

/** episode とその進捗の組。pull の応答の 1 件で、worker の Port と契約が共有する。 */
export type EpisodeProgressEntry = {
  readonly episodeId: EpisodeId;
  readonly progress: EpisodeProgress;
};
