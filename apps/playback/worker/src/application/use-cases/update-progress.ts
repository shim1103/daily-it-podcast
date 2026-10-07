import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";
import {
  loadExistingProgress,
  persistMergedProgress,
  toProgressWriteUseCaseOutput,
} from "../progress/persist-merged-progress.ts";
import type {
  ProgressWriteUseCaseInput,
  ProgressWriteUseCaseOutput,
} from "../progress/progress-write-use-case-io.ts";

/**
 * 進捗 update（HTTP PATCH progress）。途中更新・stop 時の位置同期。
 *
 * 行なしは {@link ProgressNotFoundError}。行ありは merge して upsert する
 * （Decision 2026-09-22T18-58-38 / 2026-10-01T18-54-14）。
 *
 * @require input は Controller が契約 schema の検証済み値から作った
 * @ensure ProgressWriteUseCaseOutput（勝ち側 first*）を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function updateProgress(
  repository: ProgressRepository,
  input: ProgressWriteUseCaseInput,
): Promise<ProgressWriteUseCaseOutput> {
  const existing = await loadExistingProgress(repository, input.episodeId);
  if (existing === null) {
    throw new ProgressNotFoundError(`進捗行が無い: ${input.episodeId}`);
  }
  const merged = await persistMergedProgress(repository, input.episodeId, existing, {
    positionSec: input.positionSec,
    clientAt: input.clientAt,
  });
  return toProgressWriteUseCaseOutput(merged);
}
