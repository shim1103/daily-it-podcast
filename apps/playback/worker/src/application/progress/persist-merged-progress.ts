import type { EpisodeProgress, ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";
import {
  mergeProgress,
  mergeProgressAsCompleted,
  type ProgressMergeIncoming,
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
 * create／update 用。merge 結果を upsert し、永続後の進捗本体を返す。
 */
export async function persistMergedProgress(
  repository: ProgressRepository,
  episodeId: string,
  existing: EpisodeProgress | null,
  incoming: ProgressMergeIncoming,
): Promise<EpisodeProgress> {
  const merged = mergeProgress(existing, incoming);
  await repository.upsertProgress({ episodeId, ...merged });
  return merged;
}

/**
 * complete 用。既存行必須の完走 merge を upsert する。
 *
 * @require existing は呼び出し側が行ありを確認済み
 */
export async function persistCompletedProgress(
  repository: ProgressRepository,
  episodeId: string,
  existing: EpisodeProgress,
  incoming: ProgressMergeIncoming,
): Promise<EpisodeProgress> {
  const merged = mergeProgressAsCompleted(existing, incoming);
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
