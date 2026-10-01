import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/** A 足場の sentinel。意味のある first* は C（行なし 404・merge）が決める。 */
const SENTINEL_WRITE_RESPONSE: ProgressWriteResponse = {
  firstPlayedAt: "1970-01-01T00:00:00.000Z",
  firstCompletedAt: null,
};

/**
 * 進捗 update（HTTP PATCH progress）。途中更新・stop 時の位置同期。
 *
 * A: Port upsert へ薄い委譲。行なし 404・merge は C。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function updateProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  // todo: 行なし 404・merge を Application に置き、sentinel upsert を消す（concern 5）
  await repository.upsertProgress({
    episodeId: command.episodeId,
    positionSec: command.positionSec,
    firstPlayedAt: SENTINEL_WRITE_RESPONSE.firstPlayedAt,
    firstCompletedAt: SENTINEL_WRITE_RESPONSE.firstCompletedAt,
    lastPlayedAt: command.clientAt,
  });
  return SENTINEL_WRITE_RESPONSE;
}
