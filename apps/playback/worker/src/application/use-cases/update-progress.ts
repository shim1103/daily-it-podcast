import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";
import {
  loadExistingProgress,
  persistMergedProgress,
  toProgressWriteResponse,
} from "../progress/persist-merged-progress.ts";
import type { ProgressWriteCommand } from "../progress/progress-write-command.ts";

/**
 * 進捗 update（HTTP PATCH progress）。途中更新・stop 時の位置同期。
 *
 * 行なしは {@link ProgressNotFoundError}。行ありは merge して upsert する
 * （Decision 2026-09-22T18-58-38 / 2026-10-01T18-54-14）。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse（勝ち側 first*）を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function updateProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  const existing = await loadExistingProgress(repository, command.episodeId);
  if (existing === null) {
    throw new ProgressNotFoundError(`進捗行が無い: ${command.episodeId}`);
  }
  const merged = await persistMergedProgress(repository, command.episodeId, existing, {
    positionSec: command.positionSec,
    clientAt: command.clientAt,
  });
  return toProgressWriteResponse(merged);
}
