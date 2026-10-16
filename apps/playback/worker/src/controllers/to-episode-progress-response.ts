import type { EpisodeProgressResponse } from "../../../contracts/index.ts";
import type { EpisodeProgress } from "../../../domain/index.ts";

/**
 * domain の進捗を契約の応答の形へ写す。list の embed と pull が共有する。
 *
 * @ensure 契約の4 field だけを持つ新しい object を返す。domain の型に field が足されても応答へ漏らさない
 */
export function toEpisodeProgressResponse(progress: EpisodeProgress): EpisodeProgressResponse {
  return {
    positionSec: progress.positionSec,
    firstPlayedAt: progress.firstPlayedAt,
    firstCompletedAt: progress.firstCompletedAt,
    lastPlayedAt: progress.lastPlayedAt,
  };
}
