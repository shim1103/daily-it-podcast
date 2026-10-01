import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import {
  loadExistingProgress,
  persistMergedProgress,
  toProgressWriteResponse,
} from "../progress/persist-merged-progress.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/**
 * 進捗 create（HTTP POST progress）。初回 play の永続入口。
 *
 * 行なしは merge で初回行を作る。行ありは Error にせず冪等に merge する
 * （Decision 2026-09-22T18-58-38 / 2026-10-01T18-54-14）。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse（勝ち側 first*）を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function createProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  const existing = await loadExistingProgress(repository, command.episodeId);
  const merged = await persistMergedProgress(repository, command.episodeId, existing, {
    positionSec: command.positionSec,
    clientAt: command.clientAt,
  });
  return toProgressWriteResponse(merged);
}
