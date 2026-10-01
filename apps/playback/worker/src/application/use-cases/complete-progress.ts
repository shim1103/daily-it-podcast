import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/** A 足場の sentinel。完走ゾーン判定・merge は C。 */
const SENTINEL_WRITE_RESPONSE: ProgressWriteResponse = {
  firstPlayedAt: "1970-01-01T00:00:00.000Z",
  firstCompletedAt: null,
};

/**
 * 進捗 complete（HTTP POST progress/complete）。完走ゾーン突入の記録。
 *
 * A: Port upsert へ薄い委譲。ゾーン判定・merge・原稿 durationSec 取得は C。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function completeProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  // todo: 完走ゾーン・merge・getManuscript を Application に置き、sentinel upsert を消す（concern 6）
  await repository.upsertProgress({
    episodeId: command.episodeId,
    positionSec: command.positionSec,
    firstPlayedAt: SENTINEL_WRITE_RESPONSE.firstPlayedAt,
    firstCompletedAt: SENTINEL_WRITE_RESPONSE.firstCompletedAt,
    lastPlayedAt: command.clientAt,
  });
  return SENTINEL_WRITE_RESPONSE;
}
