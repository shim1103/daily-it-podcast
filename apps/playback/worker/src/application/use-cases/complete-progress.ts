import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import { ProgressRuleError } from "../../entities/errors/progress-rule-error.ts";
import { verifyManuscript } from "../manuscript/verify-manuscript.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";
import { isInCompleteZone } from "../progress/is-in-complete-zone.ts";
import {
  loadExistingProgress,
  persistCompletedProgress,
  toProgressWriteResponse,
} from "../progress/persist-merged-progress.ts";
import type { ProgressWriteCommand } from "../progress/progress-write-command.ts";

/**
 * 進捗 complete（HTTP POST progress/complete）。完走ゾーン突入の記録。
 *
 * durationSec は原稿 1 件取得から得る。ゾーン外は {@link ProgressRuleError}。
 * 行なしは {@link ProgressNotFoundError}（Decision 2026-09-22T18-58-38）。
 * 原稿欠落・不適合は {@link EpisodeContentError}（Decision 2026-10-01T18-54-14）。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse（勝ち側 first*）を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function completeProgress(
  episodeRepository: EpisodeRepository,
  progressRepository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  const raw = await episodeRepository.getManuscript(command.episodeId);
  if (raw === undefined) {
    throw new EpisodeContentError(`原稿が無い: ${command.episodeId}`);
  }
  const manuscript = verifyManuscript(raw.json, raw.stem);

  if (!isInCompleteZone(command.positionSec, manuscript.durationSec)) {
    throw new ProgressRuleError(`完走ゾーン外: ${command.episodeId}`);
  }

  const existing = await loadExistingProgress(progressRepository, command.episodeId);
  if (existing === null) {
    throw new ProgressNotFoundError(`進捗行が無い: ${command.episodeId}`);
  }

  const merged = await persistCompletedProgress(progressRepository, command.episodeId, existing, {
    positionSec: command.positionSec,
    clientAt: command.clientAt,
  });
  return toProgressWriteResponse(merged);
}
