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
 * 進捗 create（HTTP POST progress）。初回 play の永続入口。
 *
 * 行なしは merge で初回行を作る。行ありは Error にせず冪等に merge する
 * （Decision 2026-09-22T18-58-38 / 2026-10-01T18-54-14）。
 *
 * @require input は Controller が契約 schema の検証済み値から作った
 * @ensure ProgressWriteUseCaseOutput（勝ち側 first*）を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function createProgress(
  repository: ProgressRepository,
  input: ProgressWriteUseCaseInput,
): Promise<ProgressWriteUseCaseOutput> {
  const existing = await loadExistingProgress(repository, input.episodeId);
  const merged = await persistMergedProgress(repository, input.episodeId, existing, {
    positionSec: input.positionSec,
    clientAt: input.clientAt,
  });
  return toProgressWriteUseCaseOutput(merged);
}
