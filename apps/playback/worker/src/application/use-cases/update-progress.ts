import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/**
 * 進捗 update（HTTP PATCH progress）。途中更新・stop 時の位置同期。
 *
 * A: Port へ委譲するだけ。行なし 404・merge は C。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function updateProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  return repository.writeProgress(command);
}
