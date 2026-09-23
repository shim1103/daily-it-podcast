import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/**
 * 進捗 create（HTTP POST progress）。初回 play の永続入口。
 *
 * A: Port へ委譲するだけ。冪等・merge・Domain Error は C。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function createProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  return repository.writeProgress(command);
}
