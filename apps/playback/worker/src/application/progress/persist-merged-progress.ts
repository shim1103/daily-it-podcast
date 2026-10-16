import type { EpisodeId, EpisodeProgress, ProgressReport } from "../../../../domain/index.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";
import { mergeProgress, mergeProgressAsCompleted } from "./merge-progress.ts";
import type { ProgressWriteUseCaseOutput } from "./progress-write-use-case-io.ts";

/**
 * episodeId の既存進捗を 1 件読む。無い時は null。
 */
export async function loadExistingProgress(
  repository: ProgressRepository,
  episodeId: EpisodeId,
): Promise<EpisodeProgress | null> {
  const rows = await repository.getByEpisodeIds({ episodeIds: [episodeId] });
  return rows.get(episodeId) ?? null;
}

/**
 * create／update 用。merge 結果を upsert し、永続後の進捗本体を返す。
 */
export async function persistMergedProgress(
  repository: ProgressRepository,
  episodeId: EpisodeId,
  existing: EpisodeProgress | null,
  incoming: ProgressReport,
): Promise<EpisodeProgress> {
  const merged = mergeProgress(existing, incoming);
  // todo: C が getProgressWithVersion と insertProgressIfAbsent・replaceProgressIfVersion の CAS 再試行と skew 判定へ置換したら、upsertProgress の呼び出しを消す
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
  episodeId: EpisodeId,
  existing: EpisodeProgress,
  incoming: ProgressReport,
): Promise<EpisodeProgress> {
  const merged = mergeProgressAsCompleted(existing, incoming);
  // todo: C が getProgressWithVersion と insertProgressIfAbsent・replaceProgressIfVersion の CAS 再試行と skew 判定へ置換したら、upsertProgress の呼び出しを消す
  await repository.upsertProgress({ episodeId, ...merged });
  return merged;
}

/** Write の出力は勝ち側 first* だけ（positionSec は載せない）。 */
export function toProgressWriteUseCaseOutput(
  progress: EpisodeProgress,
): ProgressWriteUseCaseOutput {
  return {
    firstPlayedAt: progress.firstPlayedAt,
    firstCompletedAt: progress.firstCompletedAt,
  };
}
