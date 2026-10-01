import type { EpisodeProgress, ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";
import {
  mergeProgress,
  type ProgressMergeIncoming,
  type ProgressMergeOptions,
} from "./merge-progress.ts";

/**
 * episodeId の既存進捗を 1 件読む。無い時は null。
 */
export async function loadExistingProgress(
  repository: ProgressRepository,
  episodeId: string,
): Promise<EpisodeProgress | null> {
  const rows = await repository.getByEpisodeIds([episodeId]);
  return rows.get(episodeId) ?? null;
}

/**
 * merge 結果を upsert し、永続後の進捗本体を返す。
 */
export async function persistMergedProgress(
  repository: ProgressRepository,
  episodeId: string,
  existing: EpisodeProgress | null,
  incoming: ProgressMergeIncoming,
  options: ProgressMergeOptions = {},
): Promise<EpisodeProgress> {
  const merged = mergeProgress(existing, incoming, options);
  await repository.upsertProgress({ episodeId, ...merged });
  return merged;
}

/** Write 応答は勝ち側 first* だけ（positionSec は載せない）。 */
export function toProgressWriteResponse(progress: EpisodeProgress): ProgressWriteResponse {
  return {
    firstPlayedAt: progress.firstPlayedAt,
    firstCompletedAt: progress.firstCompletedAt,
  };
}
