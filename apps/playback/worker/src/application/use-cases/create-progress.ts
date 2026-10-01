import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/** A 足場の sentinel。意味のある first* は C（merge・冪等）が決める。 */
const SENTINEL_WRITE_RESPONSE: ProgressWriteResponse = {
  firstPlayedAt: "1970-01-01T00:00:00.000Z",
  firstCompletedAt: null,
};

/**
 * 進捗 create（HTTP POST progress）。初回 play の永続入口。
 *
 * A: Port upsert へ薄い委譲。冪等・merge・Domain Error は C。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function createProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  // todo: merge・冪等を Application に置き、sentinel upsert を消す（concern 4+）
  await repository.upsertProgress({
    episodeId: command.episodeId,
    positionSec: command.positionSec,
    firstPlayedAt: SENTINEL_WRITE_RESPONSE.firstPlayedAt,
    firstCompletedAt: SENTINEL_WRITE_RESPONSE.firstCompletedAt,
    lastPlayedAt: command.clientAt,
  });
  return SENTINEL_WRITE_RESPONSE;
}
